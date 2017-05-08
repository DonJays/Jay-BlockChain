package main 

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)


var logger = shim.NewLogger("CTACChaincode")

//===================================================================================================
//	 Structure Definitions
//===================================================================================================

//===================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//=====================================================================================================
type  SimpleChainCode struct {
}

//=====================================================================================================
//	Booking - Defines the structure for a room reservation. JSON on right tells it what JSON fields to 
//            map to that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//====================================================================================================
type Booking struct {
	//ObjectType		string `json:"docType"` //docType is used to distinguish the various types of objects in state database  not sure why its needed
	TravelAgentID	string 	`json:"TravelAgentId"`	//unique agend id
	BookingID		int	   	`json:"bookingId"` 		//booking ID
	TotalBill		float32	`json:"totalBill"`  	//no of days client will be staying		
	IsSettled		bool   	`json:"isSettled"`  	//is booking is pending or completed
	Payables		float32	`json:"payables"`   	//final payables to agent
	status			int		`json:"status"`			//Status about the booking during its life cycle
	//CheckedOutDt		string	`json:"CheckedOutDt"`	//Checked out date of the guest
	//GuestName			string	`json:"GuestName"` 		//Guest Name who stayed in the hotel
	//ArrivalDt			string	`json:"ArrivalDt"` 		//Guest Arrival Date
	//RoomRate			float32	`json:"RoomRate"` 		//Room rate which booking happen and eligilble for commission to travel agent
	//RatePgm			string 	`json:"RatePgm"` 		//RatePgm name reference for room rate 
	//CommRate			float32 `json:"CommRate"`		//Commission rate for the booking to travel agent
	//CommRev			float32 `json:"CommRev"`		//Commission revenue calcualted after guest checkout during one button process whic will paid to travel agent
	//UpaidRC			string	`json:"UpaidRC"`		//Unpaid Reason  which may updated at hotel after the batch/report generated in the system
	//AdjuRC			string	`json:"UpaidRC"`		//Adjustment Reason  which may updated at hotel after the batch/report generated in the system
}


//=======================================================================================================
//	 Status types - Booking has different staus, this is part of the business logic to determine what can 
//					be done to the Booking at points in it's lifecycle
//=======================================================================================================
const   STATE_RESERVATION  		=  0
const   STATE_INHOUSE  			=  1
const   STATE_CHECKOUT		 	=  2
const   STATE_CANCEL 			=  3



//========================================================================================================
//	 ---------        Main ------------- main - Starts up the chaincode
//========================================================================================================
func main() {

	err := shim.Start(new(SimpleChainCode))
	if err != nil { fmt.Printf("Error starting Chaincode: %s", err) }
}

//==========================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==========================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	var book Booking

	bytes, err := json.Marshal(book)
    if err != nil { return nil, errors.New("Error creating Booking record") }
	err = stub.PutState("book", bytes)

	return nil, nil
}



//=========================================================================================================
//	 	------------------		General Functions       ------------------
//=========================================================================================================

//==========================================================================================================
//	 retrieveBooking - Gets the 'booking' in the ledger then converts it from the stored
//					JSON into the booking struct for use.
//==========================================================================================================
func (t *SimpleChaincode) retrieveBooking(stub shim.ChaincodeStubInterface, args[] string) ([]byte, error) {

	var b Booking

	bytes, err := stub.GetState(args[0]);
	if err != nil {	fmt.Printf("Failed to invoke Booking ID: %s", err); return b, errors.New("Error retrieving booking with BookingID = " + BookingID) }
	err = json.Unmarshal(bytes, &v);

    if err != nil {	fmt.Printf("Invalid booking record "+string(bytes)+": %s", err); return v, errors.New("Invalid booking record"+string(bytes))	}
	return b, nil
}


/=============================================================================================================
// save_changes - Writes to the ledger the bookinh struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//============================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, b Booking) (bool, error) {

	bytes, err := json.Marshal(b)
	if err != nil { fmt.Printf("Error converting booking record: %s", err); return false, errors.New("Error converting booking record") }

	err = stub.PutState(b.bookingID, bytes)
	if err != nil { fmt.Printf("Error storing booking record: %s", err); return false, errors.New("Error storing booking record") }

	return true, nil
}



//=========================================================================================================
//	 	------------------		Router Functions       ------------------
//=========================================================================================================

//=========================================================================================================
//	Invoke - Called on chaincode invoke, entry point. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> initBooking..
//=========================================================================================================

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	
	function,args:=stub.GetFunctionAndParameters()
	fmt.Println("invoke is running" +function)
	
	caller:= args[0]
	
	//divert different functions
	if function == "initBooking" && caller == "TA" { 
		return t.initBooking(stub,args[0])
	} else if function == "updateSettled" && caller == "HOTEL" {
		return t.updateSettled(stub,args[0])
	} else if function == "updatePayables" && caller == "CTAC" {
		return t.updateTAPay(stub,args[0])
	}else {
		return nil, errors.New("Function of the name "+ function +" doesn't exist.")
	}
}


//=========================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=========================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	
    logger.Debug("function--> ", function)

	if function == "getStatus" {
		if len(args) != 1 { fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed") }
		return t.getStatus(stub,args[0])
	}
	
	return nil, errors.New("Received unknown function invocation " + function)
}



//==========================================================================================================
//	 Ping - Pings the peer to keep the connection alive
//==========================================================================================================
func (t *SimpleChaincode) ping(stub shim.ChaincodeStubInterface) ([]byte, error) {
	return []byte("Hello, I am Alive!"), nil
}

//==========================================================================================================
//	 	------------------		Update Functions       ------------------
//===========================================================================================================

/===========================================================================================================
//	initBooiking - Called on chaincode invoke. 
/===========================================================================================================

func (t *SimpleChaincode) initBooking(stub shim.ChaincodeStubInterface, function string, args [] string)  ([]byte, error) {
	
	var err error
	var b Booking
	 
	//Passing values in function arguments
	agentID := strings.ToLower(args[0])
	bookingID := strconv.Atoi(args[1])
	totalBill := args[1])
	isSettled := args[3]
	payables := args[4]
	
	
	booking_json := &booking(objectType, agentID, bookingID, totalBill, isSettled, payables)
	
	err = json.Unmarshal([]byte(bokking_json), &b)				// Convert the JSON defined above into a booking object for go
	if err != nil { return nil, errors.New("Invalid JSON object")
		}
	
	bookingAsJsonBytes,err := json.Marshal(booking_json)
	if err != nil { return shim.Error(err.Error())
	}
	
	//put the marshal into chaincode
	err = stub.putState(bookingID, bookingAsJsonBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	
	retrun nil
}


/=========================================================================================================
//	updateSettled - Called on chaincode invoke. 
/=========================================================================================================
func (t *SimpleChaincode) updateSettled(stub shim.ChaincodeStubInterface, function string, args [] string)  ([]byte, error) {
	
	// arg0 will be caller arg1 booking id  arg2 totalbill
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting Booking ID")
	}
	
	bookingID := args[1]
	bill := args[2]
	
	bookingAsBytes, err := stub.getState(bookingID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + bookingID + "\"}"
		return shim.Error(jsonResp)
	} else if bookingAsBytes == nil {
		jsonResp := "{\"Error\":\"Booking ID does not exist: " + bookingID + "\"}"
		return shim.Error(jsonResp)
	} 
	
	
	bookingTemp := booking{}
	err:= json.Unmarshal(bookingAsBytes, &bookingTemp)
	
	if err != nil {
		return shim.Error(err.Error())
	}
	
	bookingTemp.IsSettled = true
	bookingTemp.Totalbill = bill
	
	bookingAsJsonBytes, err := json.Marshal(bookingTemp)
	err = stub.putState(bookingID, bookingAsJsonBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("settled status and total bill are successfully updated")
	
}

/=========================================================================================================
//	updateTAPay - Called on chaincode invoke. 
/=========================================================================================================
func(t *SimpleChaincode) updateTAPay(stub shim.ChaincodeStubInterface, function string, args [] string)  ([]byte, error) {
	// arg0 will be caller
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting Booking ID")
	}
	
	bookingID := args[1]
	bookingAsBytes, err := stub.getState(bookingID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + bookingID + "\"}"
		return shim.Error(jsonResp)
	} else if bookingAsBytes == nil {
		jsonResp := "{\"Error\":\"Booking ID does not exist: " + bookingID + "\"}"
		return shim.Error(jsonResp)
	} 
	
	bookingTemp := booking{}
	
	err:= json.Unmarshal(bookingAsBytes, &bookingTemp)
	
	if err != nil {
		return shim.Error(err.Error())
	}
	
	
	if bookingTemp.IsSettled  != true {
		return shim.Error("payables cant be updated as settlement is pending for booking id" +bookingID)
	}
	
	bookingTemp.Payables = bookingTemp.TotalBill *0.02
	
	bookingAsJsonBytes, err := json.Marshal(bookingTemp)
	err = stub.putState(bookingID, bookingAsJsonBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("payables are updated are successfully updated")
	}

//=========================================================================================================
//	 	------------------		Read Functions       ------------------
//=========================================================================================================

/===========================================================================================================
//	getStatus - Called on chaincode invoke. 
/===========================================================================================================
func(t *SimpleChaincode) getStatus(stub shim.ChaincodeStubInterface, function string, args [] string)  ([]byte, error) {

// arg0 will be caller
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting Booking ID")
	}
	
	bookingID := args[1]
	bookingAsBytes, err := stub.getState(bookingID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + bookingID + "\"}"
		return shim.Error(jsonResp)
	} else if bookingAsBytes == nil {
		jsonResp := "{\"Error\":\"Booking ID does not exist: " + bookingID + "\"}"
		return shim.Error(jsonResp)
	} 
	
	bookingTemp := booking{}
	
	err:= json.Unmarshal(bookingAsBytes, &bookingTemp)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.retrieveBooking(bookingID, bookingAsJsonBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("payables are updated are successfully updated")
	}
	
}


