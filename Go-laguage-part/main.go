package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const (
	TO_ALL   int    = 0x3FFF
	HUB_NAME string = "HUB01"
	HOST     string = "localhost"
	PORT     string = "9998"
	TYPE     string = "http"
)

// ENUM для команд протокола (cmd)
type CMD byte

// Команды протокола (cmd)
const (
	WHOISHERE CMD = 0x01
	IAMHERE   CMD = 0x02
	GETSTATUS CMD = 0x03
	STATUS    CMD = 0x04
	SETSTATUS CMD = 0x05
	TICK      CMD = 0x06
)

// ENUM для устройств в сети
type DEV_TYPE byte

// Типы устройств
const (
	HUB    DEV_TYPE = 0x01
	SENSOR DEV_TYPE = 0x02
	SWITCH DEV_TYPE = 0x03
	LAMP   DEV_TYPE = 0x04
	SOCKET DEV_TYPE = 0x05
	CLOCK  DEV_TYPE = 0x06
)

type Packets []Packet

func packetsFromBytes(bytes []byte) *Packets {
	n := len(bytes)
	skip := 0
	var pcts Packets
	for skip < n {
		curPacket, newSkip := packetFromBytes(bytes[skip:])
		skip += newSkip
		pcts = append(pcts, *curPacket)
	}
	return &pcts
}

func (pcts Packets) toBytes() []byte {
	byteArr := make([]byte, 0)
	for _, pct := range pcts {
		byteArr = append(byteArr, pct.toBytes()...)
	}
	return byteArr
}

// Packet — пакет, передаваемый по сети
type Packet struct {
	Length  byte    `json:"length"`  // размер поля payload в октетах (байтах);
	Payload Payload `json:"payload"` //  данные, передаваемые в пакете, конкретный формат данных для каждого типа пакета;
	Crc8    byte    `json:"crc8"`    // контрольная сумма поля payload, вычисленная по алгоритму cyclic redundancy check 8
}

func packetFromBytes(bytes []byte) (*Packet, int) {
	payloadLength := bytes[0]
	payload := bytes[1 : payloadLength+1]
	crc8 := bytes[payloadLength+1]
	crc8comp := ComputeCRC8Simple(payload)

	if crc8 != crc8comp {
		log.Fatal("control sums unequal")
	}
	// Раскодировать если совпадает crc8,
	// в структуру, которая будет выглядеть уже так
	pld := payloadFromBytes(payload)
	pct := Packet{
		Length:  payloadLength,
		Payload: *pld,
		Crc8:    crc8,
	}
	return &pct, int(payloadLength) + 2
}

func (pct Packet) toBytes() []byte {
	byteArr := make([]byte, 0)
	byteArr = append(byteArr, pct.Length)
	byteArr = append(byteArr, pct.Payload.toBytes()...)
	byteArr = append(byteArr, pct.Crc8)
	return byteArr
}

type CBDer interface {
	toBytes() []byte
}

// Payload — данные, передаваемые в пакете
type Payload struct {
	Src     int      `json:"src"`                // 14-битный “адрес” устройства-отправителя;
	Dst     int      `json:"dst"`                // 14-битный “адрес” устройства-получателя, причем адреса 0x0000 и 0x3FFF (16383) зарезервированы. Адрес 0x3FFF означает “широковещательную” рассылку
	Serial  int      `json:"serial"`             // порядковый номер пакета, отправленного устройством, от момента его включения.
	DevType DEV_TYPE `json:"dev_type"`           // тип устройства, отправившего пакет;
	Cmd     CMD      `json:"cmd"`                // команда протокола;
	CmdBody CBDer    `json:"cmd_body,omitempty"` // формат которого зависит от команды протокола
}

func payloadFromBytes(bytes []byte) *Payload {
	srcULEB, skip1 := decodeULEB128(bytes)
	dstULEB, skip2 := decodeULEB128(bytes[skip1:])
	serialULEB, skip3 := decodeULEB128(bytes[skip1+skip2:])
	skip := skip1 + skip2 + skip3
	pld := Payload{
		Src:     srcULEB,
		Dst:     dstULEB,
		Serial:  serialULEB,
		DevType: DEV_TYPE(bytes[skip]),
		Cmd:     CMD(bytes[skip+1]),
		CmdBody: nil,
	}

	cmdBodyBytes := bytes[skip+2:]
	cmdParsed := parseCBody(pld.DevType, pld.Cmd, cmdBodyBytes)
	if cmdParsed != nil {
		pld.CmdBody = cmdParsed
	}
	return &pld
}

func (pld Payload) toBytes() []byte {
	byteArr := make([]byte, 0)
	byteArr = append(byteArr, encodeULEB128(pld.Src)...)
	byteArr = append(byteArr, encodeULEB128(pld.Dst)...)
	byteArr = append(byteArr, encodeULEB128(pld.Serial)...)
	byteArr = append(byteArr, []byte{byte(pld.DevType), byte(pld.Cmd)}...)
	if pld.CmdBody != nil {
		byteArr = append(byteArr, (pld.CmdBody).toBytes()...)
	}
	return byteArr
}

type CBName struct {
	DevName string `json:"dev_name"`
}

func (cbn CBName) toBytes() []byte {
	byteArr := []byte{byte(len(cbn.DevName))}
	return append(byteArr, []byte(cbn.DevName)...)
}

type CBSensors struct {
	DevName  string         `json:"dev_name"`
	DevProps EnvSensorProps `json:"dev_props"`
}

type EnvSensorProps struct {
	Sensors  byte      `json:"sensors"`
	Triggers []Trigger `json:"triggers"`
}

type Trigger struct {
	Op    byte   `json:"op"`
	Value int    `json:"value"`
	Name  string `json:"name"`
}

func (cbn CBSensors) toBytes() []byte {
	// DevName (len, bytes) - Sensors (1 byte) - trigger len - for {1 byte - encode ULEB - string (len, bytes)}
	nameLen := byte(len(cbn.DevName))
	name := []byte(cbn.DevName)
	sensors := cbn.DevProps.Sensors
	triggerLen := len(cbn.DevProps.Triggers)
	byteArr := []byte{nameLen}
	byteArr = append(byteArr, name...)
	byteArr = append(byteArr, []byte{sensors, byte(triggerLen)}...)
	for i := 0; i < triggerLen; i++ {
		curBytes := []byte{cbn.DevProps.Triggers[i].Op}
		curBytes = append(curBytes, encodeULEB128(cbn.DevProps.Triggers[i].Value)...)
		curBytes = append(curBytes, byte(len(cbn.DevProps.Triggers[i].Name)))
		curBytes = append(curBytes, []byte(cbn.DevProps.Triggers[i].Name)...)
		byteArr = append(byteArr, curBytes...)
	}
	return byteArr
}

type CBSensor struct {
	Values []int `json:"values"`
}

func (cbs CBSensor) toBytes() []byte {
	// Vals (len, bytes)
	byteArrLen := byte(len(cbs.Values))
	byteArr := []byte{byteArrLen}
	for _, val := range cbs.Values {
		byteArr = append(byteArr, encodeULEB128(val)...)
	}
	return byteArr
}

type CBSwitch struct {
	DevName  string   `json:"dev_name"`
	DevProps DevProps `json:"dev_props"`
}

type DevProps struct {
	DevNames []string `json:"dev_names"`
}

func (cbs CBSwitch) toBytes() []byte {
	// DevName (len, bytes) - len str_slice - for {string (len, bytes)}
	devNameLen := byte(len(cbs.DevName))
	devName := []byte(cbs.DevName)
	devPropsLen := byte(len(cbs.DevProps.DevNames))
	byteArr := []byte{devNameLen}
	byteArr = append(byteArr, devName...)
	byteArr = append(byteArr, devPropsLen)
	for _, curDevName := range cbs.DevProps.DevNames {
		curDevNameLen := byte(len(curDevName))
		byteArr = append(byteArr, curDevNameLen)
		byteArr = append(byteArr, []byte(curDevName)...)
	}
	return byteArr
}

type CBValue struct {
	Value byte `json:"value"`
}

func (cbv CBValue) toBytes() []byte {
	return []byte{cbv.Value}
}

type CBTimestamp struct {
	Timestamp int `json:"timestamp"`
}

func (cbt CBTimestamp) toBytes() []byte {
	return encodeULEB128(cbt.Timestamp)
}

func parseCBody(device DEV_TYPE, cmd CMD, cmdBodyBytes []byte) CBDer {
	switch {
	case (device == HUB || device == SOCKET || device == LAMP || device == CLOCK) && (cmd == WHOISHERE || cmd == IAMHERE):
		nameLength := cmdBodyBytes[0]
		name := cmdBodyBytes[1 : nameLength+1]
		return CBName{string(name)}
	case device == SENSOR && (cmd == WHOISHERE || cmd == IAMHERE):
		nameLength := cmdBodyBytes[0]
		name := cmdBodyBytes[1 : nameLength+1]
		sensors := cmdBodyBytes[nameLength+1]
		triggerLength := cmdBodyBytes[nameLength+2]
		triggers := make([]Trigger, triggerLength)
		curSkip := int(nameLength) + 3
		for i := 0; i < int(triggerLength); i++ {
			curOp := cmdBodyBytes[curSkip]
			curSkip++
			curVal, skipULEB := decodeULEB128(cmdBodyBytes[curSkip:])
			curNameLen := cmdBodyBytes[curSkip+skipULEB]
			curName := string(cmdBodyBytes[curSkip+skipULEB+1 : curSkip+skipULEB+1+int(curNameLen)])
			curSkip += skipULEB + int(curNameLen) + 1
			curTrigger := Trigger{
				Op:    curOp,
				Value: curVal,
				Name:  curName,
			}
			triggers[i] = curTrigger
		}
		return CBSensors{
			DevName: string(name),
			DevProps: EnvSensorProps{
				Sensors:  sensors,
				Triggers: triggers,
			},
		}
	case cmd == GETSTATUS && (device == SENSOR || device == SWITCH || device == LAMP || device == SOCKET):
		return nil
	case device == SENSOR && cmd == STATUS:
		valuesSize := int(cmdBodyBytes[0])
		values := make([]int, valuesSize)
		curSkip := 1
		for i := 0; i < valuesSize; i++ {
			curVal, skipULEB := decodeULEB128(cmdBodyBytes[curSkip:])
			values[i] = curVal
			curSkip += skipULEB
		}
		return CBSensor{Values: values}
	case device == SWITCH && (cmd == WHOISHERE || cmd == IAMHERE):
		nameLength := cmdBodyBytes[0]
		name := cmdBodyBytes[1 : nameLength+1]
		devNamesLen := int(cmdBodyBytes[nameLength+1])
		devNames := make([]string, devNamesLen)
		curSkip := nameLength + 2
		for i := 0; i < devNamesLen; i++ {
			curNameLen := cmdBodyBytes[curSkip]
			curName := cmdBodyBytes[curSkip+1 : curSkip+1+curNameLen]
			devNames[i] = string(curName)
			curSkip += curNameLen + 1
		}
		return CBSwitch{
			DevName:  string(name),
			DevProps: DevProps{DevNames: devNames},
		}
	case (cmd == STATUS && (device == SWITCH || device == LAMP || device == SOCKET)) ||
		(cmd == SETSTATUS && (device == LAMP || device == SOCKET)):
		value := cmdBodyBytes[0]
		return CBValue{Value: value}
	case device == CLOCK && cmd == TICK:
		timestamp, _ := decodeULEB128(cmdBodyBytes[:])
		return CBTimestamp{Timestamp: timestamp}
	}
	return nil
}

func getConnectiongString(commUrl string) string {
	if commUrl == "" {
		return fmt.Sprintf("%s://%s:%s", TYPE, HOST, PORT)
	}
	return fmt.Sprintf(commUrl)
}

func encodeULEB128(value int) []byte {
	var res []byte
	for {
		bt := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			bt |= 0x80
		}
		res = append(res, bt)
		if value == 0 {
			break
		}
	}
	return res
}

func decodeULEB128(bytes []byte) (int, int) {
	res := 0
	shift := 0
	bytesParsed := 0
	for _, bt := range bytes {
		bytesParsed++
		res |= (int(bt) & 0x7f) << shift
		shift += 7
		if bt&0x80 == 0 {
			break
		}
	}
	return res, bytesParsed
}

func ComputeCRC8Simple(bytes []byte) byte {
	const mask byte = 0x1D
	crc8 := byte(0)
	for _, curByte := range bytes {
		crc8 ^= curByte
		for i := 0; i < 8; i++ {
			if (crc8 & 0x80) != 0 {
				crc8 = (crc8 << 1) ^ mask
			} else {
				crc8 <<= 1
			}
		}
	}
	return crc8
}

func requestServer(commUrl, reqString string) ([]byte, int, error) {
	client := &http.Client{}
	req := new(http.Request)
	var err error
	if reqString == "" {
		req, err = http.NewRequest(
			http.MethodPost, getConnectiongString(commUrl), nil,
		)
	} else {
		req, err = http.NewRequest(
			http.MethodPost, getConnectiongString(commUrl), strings.NewReader(reqString),
		)
	}
	if err != nil {
		return []byte{}, http.StatusBadRequest, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, http.StatusBadRequest, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	code := resp.StatusCode
	if err != nil {
		return []byte{}, http.StatusBadRequest, err
	}
	return body, code, nil
}

func removeSpaces(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// Device - вспомогательная структура для  БД описывающая устройсво в сети
type Device struct {
	Address      int       `json:"address"`
	DevName      string    `json:"dev_name"`
	DevType      DEV_TYPE  `json:"dev_type"`
	IsOn         bool      `json:"status"`
	IsPresent    bool      `json:"is_present"`
	ConnDevs     []string  `json:"conn_devs"`
	SensorValues []int     `json:"sensor_values"`
	Sensors      byte      `json:"sensors"`
	Triggers     []Trigger `json:"triggers"`
	AnswerTime   int       `json:"answer_time"`
}

func findTime(pcts *Packets) int {
	for _, pct := range *pcts {
		if pct.Payload.DevType == CLOCK && pct.Payload.Cmd == TICK {
			clockBody := pct.Payload.CmdBody.(CBTimestamp)
			return clockBody.Timestamp
		}
	}
	return -1
}

func setState(pcts *Packets, database map[int]*Device, devs []string, state byte, src int, serial *int) {
	for _, item := range database {
		name := item.DevName
		for _, dev := range devs {
			if name == dev {
				var cmdBody CBDer = CBValue{Value: state}
				newPacket := Packet{
					Length: 0,
					Payload: Payload{
						Src:     src,
						Dst:     item.Address,
						Serial:  *serial,
						DevType: item.DevType,
						Cmd:     SETSTATUS,
						CmdBody: cmdBody,
					},
					Crc8: 0,
				}
				*serial++
				newPacket.Crc8 = ComputeCRC8Simple(newPacket.Payload.toBytes())
				newPacket.Length = byte(len(newPacket.Payload.toBytes()))
				*pcts = append(*pcts, newPacket)
				break
			}
		}
	}
}

func pingSwitches(pcts *Packets, database map[int]*Device, src int, serial *int) {
	for _, dev := range database {
		if dev.DevType == SWITCH && dev.IsPresent {
			newPacket := Packet{
				Length: 0,
				Payload: Payload{
					Src:     src,
					Dst:     dev.Address,
					Serial:  *serial,
					DevType: HUB,
					Cmd:     GETSTATUS,
					CmdBody: nil,
				},
				Crc8: 0,
			}
			*serial++
			newPacket.Crc8 = ComputeCRC8Simple(newPacket.Payload.toBytes())
			newPacket.Length = byte(len(newPacket.Payload.toBytes()))
			*pcts = append(*pcts, newPacket)
		}
	}
}

func handleResponse(database map[int]*Device, requestTimes map[int][]int, pcts, tasks *Packets, src int, serial *int) {
	answerTime := findTime(pcts)
	for _, pct := range *pcts {
		val, ok := database[pct.Payload.Src]
		if ok && !val.IsPresent && pct.Payload.Cmd != WHOISHERE {
			continue
		}
		switch pct.Payload.Cmd {
		case IAMHERE:
			dvt := pct.Payload.DevType
			adr := pct.Payload.Src
			isAlive := answerTime-requestTimes[TO_ALL][0] <= 300
			switch {
			case dvt == SWITCH:
				body := pct.Payload.CmdBody.(CBSwitch)
				database[adr] = &Device{
					Address:      adr,
					DevName:      body.DevName,
					DevType:      pct.Payload.DevType,
					IsOn:         false,
					IsPresent:    isAlive,
					ConnDevs:     body.DevProps.DevNames,
					SensorValues: nil,
					Sensors:      0,
					Triggers:     nil,
					AnswerTime:   answerTime,
				}
			case dvt == SENSOR:
				body := pct.Payload.CmdBody.(CBSensors)
				database[adr] = &Device{
					Address:      adr,
					DevName:      body.DevName,
					DevType:      pct.Payload.DevType,
					IsOn:         false,
					IsPresent:    isAlive,
					ConnDevs:     nil,
					SensorValues: nil,
					Sensors:      body.DevProps.Sensors,
					Triggers:     body.DevProps.Triggers,
					AnswerTime:   answerTime,
				}
			default:
				body := pct.Payload.CmdBody.(CBName)
				database[adr] = &Device{
					Address:      adr,
					DevName:      body.DevName,
					DevType:      pct.Payload.DevType,
					IsOn:         false,
					IsPresent:    isAlive,
					ConnDevs:     nil,
					SensorValues: nil,
					Sensors:      0,
					Triggers:     nil,
					AnswerTime:   answerTime,
				}
			}
		case WHOISHERE:
			var cmdBody CBDer = CBName{DevName: HUB_NAME}
			newPacket := Packet{
				Length: 0,
				Payload: Payload{
					Src:     src,
					Dst:     TO_ALL,
					Serial:  *serial,
					DevType: HUB,
					Cmd:     IAMHERE,
					CmdBody: cmdBody,
				},
				Crc8: 0,
			}
			*serial++
			newPacket.Crc8 = ComputeCRC8Simple(newPacket.Payload.toBytes())
			newPacket.Length = byte(len(newPacket.Payload.toBytes()))
			*tasks = append(*tasks, newPacket)

			dvt := pct.Payload.DevType
			adr := pct.Payload.Src
			switch {
			case dvt == SWITCH:
				body := pct.Payload.CmdBody.(CBSwitch)
				database[adr] = &Device{
					Address:      adr,
					DevName:      body.DevName,
					DevType:      pct.Payload.DevType,
					IsOn:         false,
					IsPresent:    true,
					ConnDevs:     body.DevProps.DevNames,
					SensorValues: nil,
					Sensors:      0,
					Triggers:     nil,
					AnswerTime:   answerTime,
				}
			case dvt == SENSOR:
				body := pct.Payload.CmdBody.(CBSensors)
				database[adr] = &Device{
					Address:      adr,
					DevName:      body.DevName,
					DevType:      pct.Payload.DevType,
					IsOn:         false,
					IsPresent:    true,
					ConnDevs:     nil,
					SensorValues: nil,
					Sensors:      body.DevProps.Sensors,
					Triggers:     body.DevProps.Triggers,
					AnswerTime:   answerTime,
				}
			default:
				body := pct.Payload.CmdBody.(CBName)
				database[adr] = &Device{
					Address:      adr,
					DevName:      body.DevName,
					DevType:      pct.Payload.DevType,
					IsOn:         false,
					IsPresent:    true,
					ConnDevs:     nil,
					SensorValues: nil,
					Sensors:      0,
					Triggers:     nil,
					AnswerTime:   answerTime,
				}
			}
		case STATUS:
			if pct.Payload.Src != TO_ALL {
				if len(requestTimes[pct.Payload.Src]) >= 2 {
					requestTimes[pct.Payload.Src] = requestTimes[pct.Payload.Src][1:]
				} else {
					delete(requestTimes, pct.Payload.Src)
				}
			}
			switch pct.Payload.DevType {
			case LAMP:
				cbv := pct.Payload.CmdBody.(CBValue)
				if cbv.Value == 1 {
					database[pct.Payload.Src].IsOn = true
				} else {
					database[pct.Payload.Src].IsOn = false
				}
			case SOCKET:
				cbv := pct.Payload.CmdBody.(CBValue)
				if cbv.Value == 1 {
					database[pct.Payload.Src].IsOn = true
				} else {
					database[pct.Payload.Src].IsOn = false
				}
			case SWITCH:
				cbv := pct.Payload.CmdBody.(CBValue)
				switch cbv.Value {
				case 1:
					database[pct.Payload.Src].IsOn = true
					devNames2TurnOn := database[pct.Payload.Src].ConnDevs
					setState(tasks, database, devNames2TurnOn, 1, src, serial)
				default:
					database[pct.Payload.Src].IsOn = false
					devNames2TurnOff := database[pct.Payload.Src].ConnDevs
					setState(tasks, database, devNames2TurnOff, 0, src, serial)
				}
			case SENSOR:
				values := pct.Payload.CmdBody.(CBSensor).Values
				database[pct.Payload.Src].SensorValues = values
				valuesAll := [4]int{-1, -1, -1, -1}
				envSensor := database[pct.Payload.Src]
				sensorTypeMask := envSensor.Sensors
				idx := 0
				for i := 0; i < 4; i++ {
					if sensorTypeMask&1 == 1 {
						valuesAll[i] = values[idx]
						idx++
					}
					sensorTypeMask = sensorTypeMask >> 1
				}
				triggers := envSensor.Triggers
				for _, trigger := range triggers {
					thresh := trigger.Value
					device := trigger.Name
					opBits := trigger.Op

					state := opBits & 1
					opBits = opBits >> 1
					greaterThen := opBits & 1
					opBits = opBits >> 1
					sensorType := opBits

					if greaterThen == 1 {
						if valuesAll[sensorType] > thresh {
							setState(tasks, database, []string{device}, state, src, serial)
						}
					} else {
						if valuesAll[sensorType] < thresh && valuesAll[sensorType] != -1 {
							setState(tasks, database, []string{device}, state, src, serial)
						}
					}
				}
			}
		}
	}
}

func simulateServer() {
	args := os.Args[1:]
	if len(args) < 2 {
		os.Exit(99)
	}
	commUrl := args[0]
	hubAddress, err := strconv.ParseInt(args[1], 16, 64)
	if err != nil {
		os.Exit(99)
	}

	database := make(map[int]*Device)
	requestTimes := make(map[int][]int)

	serialCounter := 1
	var statusCode int
	var reqStr string
	var respRawBytes, respRawBytesTrimmed, respBytes []byte

	pendingTasks := Packets{}
	var hubTime int

	for {
		var cbn CBDer = CBName{DevName: HUB_NAME}
		pcts := Packets{
			Packet{
				Length: 0,
				Payload: Payload{
					Src:     int(hubAddress),
					Dst:     TO_ALL,
					Serial:  serialCounter,
					DevType: HUB,
					Cmd:     WHOISHERE,
					CmdBody: cbn,
				},
				Crc8: 0,
			},
		}
		serialCounter++
		pcts[0].Length = byte(len(pcts[0].Payload.toBytes()))
		pcts[0].Crc8 = ComputeCRC8Simple(pcts[0].Payload.toBytes())
		reqStr = base64.RawURLEncoding.EncodeToString(pcts.toBytes())
		respRawBytes, statusCode, err = requestServer(commUrl, reqStr)
		if err != nil {
			os.Exit(99)
		}

		if statusCode == http.StatusOK {
			respRawBytesTrimmed = []byte(removeSpaces(string(respRawBytes)))
			respBytes, err = base64.RawURLEncoding.DecodeString(string(respRawBytesTrimmed))
			if err != nil {
				continue
			}
			respPcts := packetsFromBytes(respBytes)
			hubTime = findTime(respPcts)
			requestTimes[TO_ALL] = []int{hubTime}
			handleResponse(database, requestTimes, respPcts, &pendingTasks, int(hubAddress), &serialCounter)
			for _, dev := range database {
				dev.IsPresent = true
			}
			break
		} else if statusCode == http.StatusNoContent {
			os.Exit(0)
		} else {
			os.Exit(99)
		}
	}

	for _, dev := range database {
		if dev.DevType == SENSOR {
			getStatusReq := Packet{
				Length: 0,
				Payload: Payload{
					Src:     int(hubAddress),
					Dst:     dev.Address,
					Serial:  serialCounter,
					DevType: HUB,
					Cmd:     GETSTATUS,
					CmdBody: nil,
				},
				Crc8: 0,
			}
			serialCounter++
			getStatusReq.Length = byte(len(getStatusReq.Payload.toBytes()))
			getStatusReq.Crc8 = ComputeCRC8Simple(getStatusReq.Payload.toBytes())
			pendingTasks = append(pendingTasks, getStatusReq)
		}
	}

	for statusCode == http.StatusOK {
		pingSwitches(&pendingTasks, database, int(hubAddress), &serialCounter)
		for _, pct := range pendingTasks {
			curCmd := pct.Payload.Cmd
			curDst := pct.Payload.Dst
			if curCmd == GETSTATUS || curCmd == SETSTATUS {
				requestTimes[curDst] = append(requestTimes[curDst], hubTime)
			}
		}
		reqStr = base64.RawURLEncoding.EncodeToString(pendingTasks.toBytes())
		pendingTasks = Packets{}

		respRawBytes, statusCode, err = requestServer(commUrl, reqStr)

		if err != nil {
			os.Exit(99)
		}
		respRawBytesTrimmed = []byte(removeSpaces(string(respRawBytes)))
		respBytes, err = base64.RawURLEncoding.DecodeString(string(respRawBytesTrimmed))
		if err != nil {
			continue
		}

		respPcts := packetsFromBytes(respBytes)

		hubTime = findTime(respPcts)

		for address, timeQueue := range requestTimes {
			if hubTime-timeQueue[0] > 300 {
				if _, ok := database[address]; ok {
					database[address].IsPresent = false
				}
				delete(requestTimes, address)
			}
		}

		handleResponse(database, requestTimes, respPcts, &pendingTasks, int(hubAddress), &serialCounter)
	}

	if statusCode == http.StatusNoContent {
		os.Exit(0)
	}
	os.Exit(99)
}

func main() {
	simulateServer()
}
