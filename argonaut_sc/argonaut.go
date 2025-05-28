package main

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io"
    "strconv"
    "strings"

    "embed"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//go:embed argonaut_test_v2.csv
var csvFile embed.FS

type SmartContract struct {
    contractapi.Contract
}

type EmissionEntry struct {
    CF         float64 `json:"cf"`
    FlowClass2 string  `json:"flow_class2"`
}


type FlowData struct {
    FlowName     string                         `json:"flow_name"`
    LCIAMethod   string                         `json:"lcia_method"`
    Measurings map[string][]EmissionEntry    	`json:"measurings"`
}


func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    data, err := csvFile.ReadFile("argonaut_test_v2.csv")
    if err != nil {
        return fmt.Errorf("failed to read embedded CSV file: %v", err)
    }

    reader := csv.NewReader(strings.NewReader(string(data)))

    _, err = reader.Read() 
    if err != nil {
        return fmt.Errorf("failed to read CSV header: %v", err)
    }

    flowMap := make(map[string]*FlowData)

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("error reading CSV: %v", err)
        }
        if len(record) < 5 {
            continue
        }

        flowName := record[0]
        lciaMethod := record[1]
        cfStr := record[2]
        flowClass1 := record[3]
        flowClass2 := record[4]

        cf, err := strconv.ParseFloat(cfStr, 64)
        if err != nil {
            cf = 0
        }

        key := flowName

        flowData, exists := flowMap[key]
        if !exists {
            flowData = &FlowData{
                FlowName:     flowName,
                LCIAMethod:   lciaMethod,
                Measurings: make(map[string][]EmissionEntry),
            }
            flowMap[key] = flowData
        }

        emission := EmissionEntry{
            FlowClass2: flowClass2,
            CF:         cf,
        }

        flowData.Measurings[flowClass1] = append(flowData.Measurings[flowClass1], emission)
    }

    stub := ctx.GetStub()

    for key, flowData := range flowMap {
        jsonBytes, err := json.Marshal(flowData)
        if err != nil {
            return fmt.Errorf("failed to marshal JSON: %v", err)
        }

        ledgerKey := key
        err = stub.PutState(ledgerKey, jsonBytes)
        if err != nil {
            return fmt.Errorf("failed to put state: %v", err)
        }
    }

    return nil
}


func (s *SmartContract) FlowExists(ctx contractapi.TransactionContextInterface, flowName string) (bool, error) {
    flowJSON, err := ctx.GetStub().GetState(flowName)
    if err != nil {
        return false, fmt.Errorf("failed to read from world state: %v", err)
    }

    return flowJSON != nil, nil
}



func (s *SmartContract) CreateFlow(
    ctx contractapi.TransactionContextInterface,
    flowName string,
    lciaMethod string,
    flowClass1 string,
    flowClass2 string,
    cfStr string,
) error {
    stub := ctx.GetStub()

    exists, err := s.FlowExists(ctx, flowName)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("flow with name %s already exists", flowName)
    }


    cf, err := strconv.ParseFloat(cfStr, 64)
    if err != nil {
        return fmt.Errorf("invalid CF value: %v", err)
    }


    emission := EmissionEntry{
        CF:         cf,
        FlowClass2: flowClass2,
    }


    flow := FlowData{
        FlowName:   flowName,
        LCIAMethod: lciaMethod,
        Measurings: map[string][]EmissionEntry{
            flowClass1: {emission},
        },
    }


    flowJSON, err := json.Marshal(flow)
    if err != nil {
        return fmt.Errorf("failed to marshal flow data: %v", err)
    }

    err = stub.PutState(flowName, flowJSON)
    if err != nil {
        return fmt.Errorf("failed to put state: %v", err)
    }

    return nil
}



func (s *SmartContract) UpdateFlow(
    ctx contractapi.TransactionContextInterface,
    flowName string,
    flowClass1 string,
    flowClass2 string,
    cfStr string,
) error {
    stub := ctx.GetStub()

    exists, err := s.FlowExists(ctx, flowName)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("flow with name %s does not exist", flowName)
    }

    data, err := stub.GetState(flowName)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }

    cf, err := strconv.ParseFloat(cfStr, 64)
    if err != nil {
        return fmt.Errorf("invalid CF value: %v", err)
    }

    var flow FlowData
    err = json.Unmarshal(data, &flow)
    if err != nil {
        return fmt.Errorf("failed to unmarshal flow data: %v", err)
    }

    newEmission := EmissionEntry{
        FlowClass2: flowClass2,
        CF:         cf,
    }

    flow.Measurings[flowClass1] = append(flow.Measurings[flowClass1], newEmission)

    updatedJSON, err := json.Marshal(flow)
    if err != nil {
        return fmt.Errorf("failed to marshal updated flow: %v", err)
    }

    err = stub.PutState(flowName, updatedJSON)
    if err != nil {
        return fmt.Errorf("failed to update flow in ledger: %v", err)
    }

    return nil
}




func (s *SmartContract) DeleteFlow(ctx contractapi.TransactionContextInterface, flowName string) error {
    stub := ctx.GetStub()

    exists, err := s.FlowExists(ctx, flowName)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("flow with name %s does not exist", flowName)
    }

    err = stub.DelState(flowName)
    if err != nil {
        return fmt.Errorf("failed to delete flow: %v", err)
    }

    return nil
}





func (s *SmartContract) GetAllFlows(ctx contractapi.TransactionContextInterface) ([]*FlowData, error) {
    stub := ctx.GetStub()

    resultsIterator, err := stub.GetStateByRange("", "")
    if err != nil {
        return nil, fmt.Errorf("failed to get state iterator: %v", err)
    }
    defer resultsIterator.Close()

    var allFlows []*FlowData

    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, fmt.Errorf("failed to iterate results: %v", err)
        }

        var flow FlowData
        err = json.Unmarshal(queryResponse.Value, &flow)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal flow data: %v", err)
        }

        allFlows = append(allFlows, &flow)
    }

    return allFlows, nil
}


func (s *SmartContract) ReadFlow(ctx contractapi.TransactionContextInterface, flowName string) (*FlowData, error) {
    exists, err := s.FlowExists(ctx, flowName)
    if err != nil {
        return nil, err
    }
    if !exists {
        return nil, fmt.Errorf("flow with name %s does not exist", flowName)
    }

    stub := ctx.GetStub()

    data, err := stub.GetState(flowName)
    if err != nil {
        return nil, fmt.Errorf("failed to read from world state: %v", err)
    }

    var flow FlowData
    err = json.Unmarshal(data, &flow)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal flow data: %v", err)
    }

    return &flow, nil
}






func main() {
    chaincode, err := contractapi.NewChaincode(&SmartContract{})
    if err != nil {
        panic(fmt.Sprintf("Error creating argonaut chaincode: %v", err))
    }

    if err := chaincode.Start(); err != nil {
        panic(fmt.Sprintf("Error starting argonaut chaincode: %v", err))
    }
}
