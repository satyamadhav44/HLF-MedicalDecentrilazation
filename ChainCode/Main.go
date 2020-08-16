package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type patient_record struct {

}

type patient_info struct {

	Patient_Id   	     string `json: "Patient_Id"`
	Patient_Name 	     string `json: "Patient_Name"`
	Patient_DOB          string `json: "Patient_DOB"`
	Patient_DOJ			 string `json: "Patient_DOJ"`
	Hopital_Id		     string `json: "Hospital_Id"`
	Pharma_Id			 string `json: "Pharma_Id"`
	Patient_Status       string `json: "Patient_Status"`
	Patient_CreationDate string `json: "Patient_CreationDate"`

}

type hospital_info struct {

	Hospital_Id         string `json: "Hospital_Id"`
	Hospital_Name       string `json: "Hospital_Name"`
	Hospital_CreationDate string `json: "Hospital_CreationDate"`
}

type pharma_info struct {

	Pharma_Id         string `json: "Pharma_Id"`
	Pharma_Name       string `json: "Pharma_Name"`
	Pharma_CreationDate string `json: "Pharma_CreationDate"`
}

type CounterNO struct {
	Counter int `json:"counter"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(patient_record))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}

}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *patient_record) Init(APIstub shim.ChaincodeStubInterface) pb.Response {

	// Initializing Patient Counter
	PatientCounterBytes, _ := APIstub.GetState("PatientCounterNO")
	if PatientCounterBytes == nil {
		var PatientCounter = CounterNO{Counter: 0}
		PatientCounterBytes, _ := json.Marshal(PatientCounter)
		err := APIstub.PutState("PatientCounterNO", PatientCounterBytes)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to Intitate Patient Counter"))
		}
	}

	// Initializing Hispoital Counter
	HospitalCounterBytes, _ := APIstub.GetState("HospitalCounterNO")
	if HospitalCounterBytes == nil {
		var HospitalCounter = CounterNO{Counter: 0}
		HospitalCounterBytes, _ := json.Marshal(HospitalCounter)
		err := APIstub.PutState("HospitalCounterNO", HospitalCounterBytes)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to Intitate Hospital Counter"))
		}
	}

	// Initializing Pharma Counter
	PharmaCounterBytes, _ := APIstub.GetState("PharmaCounterNO")
	if PharmaCounterBytes == nil {
		var PharmaCounter = CounterNO{Counter: 0}
		PharmaCounterBytes, _ := json.Marshal(PharmaCounter)
		err := APIstub.PutState("PharmaCounterNO", PharmaCounterBytes)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to Intitate Pharma Counter"))
		}
	}

	return shim.Success(nil)

}

//getCounter to the latest value of the counter based on the counter Type provided as input parameter
func getCounter(APIstub shim.ChaincodeStubInterface, CounterType string) int {
	counterAsBytes, _ := APIstub.GetState(CounterType)
	counterAsset := CounterNO{}

	json.Unmarshal(counterAsBytes, &counterAsset)
	fmt.Sprintf("Counter Current Value %d of Asset Type %s", counterAsset.Counter, CounterType)

	return counterAsset.Counter
}

//incrementCounter to the increase value of the counter based on the Asset Type provided as input parameter by 1
func incrementCounter(APIstub shim.ChaincodeStubInterface, CounterType string) int {
	counterAsBytes, _ := APIstub.GetState(CounterType)
	counterAsset := CounterNO{}

	json.Unmarshal(counterAsBytes, &counterAsset)
	counterAsset.Counter++
	counterAsBytes, _ = json.Marshal(counterAsset)

	err := APIstub.PutState(CounterType, counterAsBytes)
	if err != nil {

		fmt.Sprintf("Failed to Increment Counter")

	}
	return counterAsset.Counter
}


// GetTxTimestampChannel Function gets the Transaction time when the chain code was executed it remains same on all the peers where chaincode executes
func (t *patient_record) GetTxTimestampChannel(APIstub shim.ChaincodeStubInterface) (string, error) {
	txTimeAsPtr, err := APIstub.GetTxTimestamp()
	if err != nil {
		fmt.Printf("Returning error in TimeStamp \n")
		return "Error", err
	}
	fmt.Printf("\t returned value from APIstub: %v\n", txTimeAsPtr)
	timeStr := time.Unix(txTimeAsPtr.Seconds, int64(txTimeAsPtr.Nanos)).String()

	return timeStr, nil
}

// get History For Record
func (t *patient_record) getHistoryForPatientRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	recordKey := args[0]

	fmt.Printf("- start getHistoryForpatientRecord: %s\n", recordKey)

	resultsIterator, err := stub.GetHistoryForKey(recordKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the key/value pair
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON vehiclePart)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForRecord returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (t *patient_record) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("function is ==> :" + function)
	//action := args[0]
	//fmt.Println(" action is ==> :" + action)
	fmt.Println(args)

	
	if function == "getHistoryForPatient" {
		return t.getHistoryForPatientRecord(stub, args)
	} else if function == "enrollPatientData" {
		return t.enrollPatientData(stub, args)
	} else if function == "updatePatientData" {
		return t.updatePatientData(stub, args)
	} else if function == "enrollHospital" {
		return t.enrollHospital(stub, args)
	} else if function == "enrollPharma" {
		return t.enrollPharma(stub, args)}
	// } else if function == "query" {
	// 	return t.Query(stub, args)
	// }
	//else if action == "history" {
	// 	return t.GetHistory(stub, args)
	// }

	fmt.Println("invoke did not find func: " + function) //error

	return shim.Error("Received unknown function")
}

//enrollHospital
func (t *patient_record) enrollHospital(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	
	if len(args) != 1 {
		return shim.Error("incorrect number of arguments, required 1")
	}

	hospitalCounter := getCounter(APIstub,"HospitalCounterNO")
	hospitalCounter++

	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(APIstub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	var hospitalAsset = hospital_info{Hospital_Id: "Hospital" + strconv.Itoa(hospitalCounter), Hospital_Name: args[0], Hospital_CreationDate: txTimeAsPtr}
	


	hospitalAssetAsBytes, errMarshal := json.Marshal(hospitalAsset)

	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error in Product: %s", errMarshal))
	}

	errPut := APIstub.PutState(hospitalAsset.Hospital_Id, hospitalAssetAsBytes)

	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to create Product Asset: %s", hospitalAsset.Hospital_Id))
	}

	//TO Increment the Product Counter
	incrementCounter(APIstub, "HospitalCounterNO")

	fmt.Println("Success in creating Product Asset %v", hospitalAsset)

	return shim.Success(nil)
}

// enroll pharma
func (t *patient_record) enrollPharma(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("incorrect number of arguments, required 1")
	}

	pharmaCounter := getCounter(APIstub,"PharmaCounterNO")
	pharmaCounter++

	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(APIstub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	var pharmaAsset = pharma_info{Pharma_Id: "Pharma" + strconv.Itoa(pharmaCounter), Pharma_Name: args[0], Pharma_CreationDate: txTimeAsPtr}


	pharmaAssetAsBytes, errMarshal := json.Marshal(pharmaAsset)

	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error in Product: %s", errMarshal))
	}

	errPut := APIstub.PutState(pharmaAsset.Pharma_Id, pharmaAssetAsBytes)

	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to create Product Asset: %s", pharmaAsset.Pharma_Id))
	}

	//TO Increment the Pharma Counter
	incrementCounter(APIstub, "PharmaCounterNO")

	fmt.Println("Success in creating Product Asset %v", pharmaAsset)

	return shim.Success(nil)
}

// enroll patient
func (t *patient_record) enrollPatientData(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {

	//To check number of arguments are 6
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments, Required 6 arguments")
	}

	//To check each argument is not null
	for i := 0; i < len(args); i++ {
		if len(args[i]) <= 0 {
			return shim.Error(string(i+1) + "st argument must be a non-empty string")
		}
	}


	PatientCounter := getCounter(APIstub, "PatientCounterNO")
	PatientCounter++

	hospitalcheck,_ := APIstub.GetState(args[3])
	
	if hospitalcheck == nil {
		return shim.Error("Given Hospital ID doesnot exist in the system")
	}

	Pharmacheck,_ := APIstub.GetState(args[4])
	
	if Pharmacheck == nil {
		return shim.Error("Given Phrama ID doesnot exist in the system")
	}
	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(APIstub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	var patientAsset = patient_info{Patient_Id: "Patient" + strconv.Itoa(PatientCounter), Patient_Name: args[0], Patient_DOB: args[1], Patient_DOJ: args[2], Hopital_Id: args[3], Pharma_Id: args[4], Patient_Status: args[5] ,Patient_CreationDate: txTimeAsPtr }

	patientAssetAsBytes, errMarshal := json.Marshal(patientAsset)

	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error in Product: %s", errMarshal))
	}

	errPut := APIstub.PutState(patientAsset.Patient_Id, patientAssetAsBytes)

	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to create Product Asset: %s", patientAsset.Patient_Id))
	}

	//TO Increment the Product Counter
	incrementCounter(APIstub, "PatientCounterNO")

	fmt.Println("Success in creating Product Asset %v", patientAsset)

	return shim.Success(nil)

}

// queryPaitient Function gets the assetspatient data based on Id provided as input
func (t *patient_record) queryPatient(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, Required 1")
	}

	fmt.Println("In Query patient data")

	PatientAssetAsBytes, _ := APIstub.GetState(args[0])
	fmt.Println(PatientAssetAsBytes, "In Query Asset")
	if PatientAssetAsBytes == nil {
		return shim.Error("Could not locate Asset")

	}

	return shim.Success(PatientAssetAsBytes)
}

// update Patient Attributes
func (t *patient_record) updatePatientData(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments, Required 6")
	}

	if len(args[0]) == 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	productBytes, _ := APIstub.GetState(args[0])

	if productBytes == nil {
		return shim.Error("Cannot Find patient Asset ")
	}

	
	var patientAsset = patient_info{Patient_Id: args[0], Patient_Status: args[1]}

	patientAssetAsBytes, errMarshal := json.Marshal(patientAsset)

	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
	}

	errPut := APIstub.PutState(patientAsset.Patient_Id, patientAssetAsBytes)

	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to Update Patient record: %s", patientAsset.Patient_Id))
	}

	fmt.Println("Success in updating Patient record %v ", patientAsset)

	return shim.Success(nil)
}
