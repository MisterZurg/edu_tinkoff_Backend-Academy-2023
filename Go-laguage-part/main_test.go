//package main
//
//func TestMain() {
//
//}
//
//var tests = []struct {
//	tcName  string
//	request string
//}{
//	{"SmartHub, WHOISHERE (1, 1):", "DAH_fwEBAQVIVUIwMeE"},
//	{"SmartHub, IAMHERE (1, 2):", "DAH_fwIBAgVIVUIwMak"},
//	{"EnvSensor, WHOISHERE (2, 1): ", "OAL_fwMCAQhTRU5TT1IwMQ8EDGQGT1RIRVIxD7AJBk9USEVSMgCsjQYGT1RIRVIzCAAGT1RIRVI03Q"},
//	{"EnvSensor, IAMHERE (2, 2): ", "OAL_fwQCAghTRU5TT1IwMQ8EDGQGT1RIRVIxD7AJBk9USEVSMgCsjQYGT1RIRVIzCAAGT1RIRVI09w"},
//	{"EnvSensor, GETSTATUS (2, 3): ", "BQECBQIDew"},
//	{"EnvSensor, STATUS (2, 4): ", "EQIBBgIEBKUB4AfUjgaMjfILrw"},
//	{"Switch, WHOISHERE (3, 1): ", "IgP_fwcDAQhTV0lUQ0gwMQMFREVWMDEFREVWMDIFREVWMDO1"},
//	{"Switch, IAMHERE (3, 2): ", "IgP_fwgDAghTV0lUQ0gwMQMFREVWMDEFREVWMDIFREVWMDMo"},
//	{"Switch, GETSTATUS (3, 3): ", "BQEDCQMDoA"},
//	{"Switch, STATUS (3, 4): ", "BgMBCgMEAac"},
//	{"Lamp, WHOISHERE (4, 1): ", "DQT_fwsEAQZMQU1QMDG8"},
//	{"Lamp, IAMHERE (4, 2): ", "DQT_fwwEAgZMQU1QMDGU"},
//	{"Lamp, GETSTATUS (4, 3): ", "BQEEDQQDqw"},
//	{"Lamp, STATUS (4, 4): ", "BgQBDgQEAaw"},
//	{"Lamp, SETSTATUS (4, 5): ", "BgEEDwQFAeE"},
//	{"Socket, WHOISHERE (5, 1): ", "DwX_fxAFAQhTT0NLRVQwMQ4"},
//	{"Socket, IAMHERE (5, 2): ", "DwX_fxEFAghTT0NLRVQwMc0"},
//	{"Socket, GETSTATUS (5, 3): ", "BQEFEgUD5A"},
//	{"Socket, STATUS (5, 4): ", "BgUBEwUEAQ8"},
//	{"Socket, SETSTATUS (5, 5): ", "BgEFFAUFAQc"},
//	{"Clock, IAMHERE (6, 2): ", "Dgb_fxUGAgdDTE9DSzAxsw"},
//	{"Clock, TICK (6, 6): ", "DAb_fxgGBpabldu2NNM"},
//}
//
////func testCase(req string) error {
////	cmd := fmt.Sprintf("echo \"%s\" | ../smart-home-binary/smarthome-linux-amd64-0.2.1 -B", req)
////	trueAns, err := exec.Command("bash", "-c", cmd).Output()
////	if err != nil {
////		return err
////	}
////
////	reqTrimmed := removeEscapeSymbols(req)
////	data, err := base64.RawURLEncoding.DecodeString(reqTrimmed)
////	if err != nil {
////		return err
////	}
////	pcts := packetsFromBytes(data)
////
////	a := removeEscapeSymbols(string(trueAns))
////	a = a[1 : len(a)-1]
////	var b string
////
////	for _, pct := range *pcts {
////		jsonPacket, err := json.Marshal(pct)
////		if err != nil {
////			return err
////		}
////		b += "," + string(jsonPacket)
////	}
////
////	b = b[1:]
////	//fmt.Println("WANT:", a)
////	//fmt.Println("HAVE:", b)
////
////	if a != b {
////		fmt.Println("WANT:", a)
////		fmt.Println("HAVE:", b)
////		return fmt.Errorf("answer is not correct")
////	}
////
////	if string(data) != string(pcts.toBytes()) {
////		fmt.Println(data)
////		fmt.Println(pcts.toBytes())
////		return fmt.Errorf("method toBytes() not performing properly")
////	}
////
////	return nil
////}
////
////func testAll() {
////	for _, test := range INPUTTESTS {
////		input := test.Request
////		err := testCase(input)
////		if err != nil {
////			fmt.Printf("\n%+v\n", test)
////			log.Fatal(err)
////		}
////	}
////	fmt.Println("All tests passed")
////}
package main

import "testing"

func TestConnect(t *testing.T) {

}
