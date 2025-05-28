package main

import (
  "fmt"
  "encoding/json"
  "log"
  "strings"
  "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
  contractapi.Contract
}

type Rating struct {
  Comment   string `json:"Comment"`
  ID        string `json:"ID"`
  ProductID string `json:"ProductID"`
  Score     int    `json:"Score"`
}

type Product struct {
  Name        string   `json:"Name"`
  Price       float64  `json:"Price"`
  ProductID   string   `json:"ProductID"`
  Quantity    int      `json:"Quantity"`
  Ratings     []Rating `json:"Ratings"`
  SellerID  string   `json:"SellerID"`
  Sold      int      `json:"Sold"`
}



func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
  products := []Product{
    {
      ProductID: "product1",
      Name:      "Wireless Mouse",
      Price:     25.99,
      Quantity:  150,
      Ratings: []Rating{
        {Comment: "Works perfectly, very responsive.", ID: "rating1", ProductID: "product1", Score: 5},
        {Comment: "Good value for the price.", ID: "rating2", ProductID: "product1", Score: 4},
      },
      SellerID: "seller1",
      Sold:     0,
    },
    {
      ProductID: "product2",
      Name:      "Yoga Mat",
      Price:     40.00,
      Quantity:  80,
      Ratings: []Rating{
        {Comment: "Comfortable and non-slip.", ID: "rating3", ProductID: "product2", Score: 5},
        {Comment: "A bit thin but still good.", ID: "rating4", ProductID: "product2", Score: 3},
        {Comment: "Colors are vibrant.", ID: "rating5", ProductID: "product2", Score: 4},
      },
      SellerID: "seller2",
      Sold:     0,
    },
    {
      ProductID: "product3",
      Name:      "Electric Kettle",
      Price:     60.50,
      Quantity:  50,
      Ratings: []Rating{
        {Comment: "Heats water quickly.", ID: "rating6", ProductID: "product3", Score: 5},
        {Comment: "Build quality could be better.", ID: "rating7", ProductID: "product3", Score: 3},
      },
      SellerID: "seller3",
      Sold:     5,
    },
    {
      ProductID: "product4",
      Name:      "Bluetooth Speaker",
      Price:     89.99,
      Quantity:  100,
      Ratings: []Rating{
        {Comment: "Great sound quality and portable.", ID: "rating8", ProductID: "product4", Score: 5},
        {Comment: "Battery life is short.", ID: "rating9", ProductID: "product4", Score: 3},
      },
      SellerID: "seller1",
      Sold:     0,
    },
    {
      ProductID: "product5",
      Name:      "Desk Lamp",
      Price:     30.00,
      Quantity:  200,
      Ratings: []Rating{
        {Comment: "Bright and adjustable.", ID: "rating10", ProductID: "product5", Score: 4},
      },
      SellerID: "seller1",
      Sold:     110,
    },
    {
      ProductID: "product6",
      Name:      "Running Shoes",
      Price:     120.00,
      Quantity:  75,
      Ratings: []Rating{
        {Comment: "Very comfortable for long runs.", ID: "rating11", ProductID: "product6", Score: 5},
        {Comment: "A bit pricey but worth it.", ID: "rating12", ProductID: "product6", Score: 4},
      },
      SellerID: "seller2",
      Sold:     15,
    },
    {
      ProductID: "product7",
      Name:      "Smart Watch",
      Price:     200.00,
      Quantity:  60,
      Ratings: []Rating{
        {Comment: "Excellent features and easy to use.", ID: "rating13", ProductID: "product7", Score: 5},
        {Comment: "Screen is a bit small.", ID: "rating14", ProductID: "product7", Score: 4},
      },
      SellerID: "seller2",
      Sold:     20,
    },
    {
      ProductID: "product8",
      Name:      "Noise Cancelling Headphones",
      Price:     250.00,
      Quantity:  40,
      Ratings: []Rating{
        {Comment: "Outstanding noise cancellation.", ID: "rating15", ProductID: "product8", Score: 5},
        {Comment: "A bit heavy but comfortable.", ID: "rating16", ProductID: "product8", Score: 4},
        {Comment: "Battery lasts long.", ID: "rating17", ProductID: "product8", Score: 5},
      },
      SellerID: "seller3",
      Sold:     10,
    },
  }

  for _, product := range products {
    productJSON, err := json.Marshal(product)
    if err != nil {
      return err
    }

    err = ctx.GetStub().PutState(product.ProductID, productJSON)
    if err != nil {
      return fmt.Errorf("failed to put to world state. %v", err)
    }
  }

  return nil
}



func (s *SmartContract) CreateProduct(ctx contractapi.TransactionContextInterface, productID string, name string, price float64, quantity int, sellerID string) error {
    exists, err := s.ProductExists(ctx, productID)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("the product %s already exists", productID)
    }

    product := Product{
        ProductID: productID,
        Name:      name,
        Price:     price,
        Quantity:  quantity,
        Ratings:   []Rating{}, // initially empty
        SellerID:  sellerID,
        Sold:      0, // inicijalno nijedan prodat proizvod
    }
    productJSON, err := json.Marshal(product)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(productID, productJSON)
}



func (s *SmartContract) ReadProduct(ctx contractapi.TransactionContextInterface, productID string) (*Product, error) {
  productJSON, err := ctx.GetStub().GetState(productID)
  if err != nil {
    return nil, fmt.Errorf("failed to read from world state: %v", err)
  }
  if productJSON == nil {
    return nil, fmt.Errorf("the product %s does not exist", productID)
  }

  var product Product
  err = json.Unmarshal(productJSON, &product)
  if err != nil {
    return nil, err
  }

  return &product, nil
}




func (s *SmartContract) UpdateProduct(ctx contractapi.TransactionContextInterface, productID string, name string, price float64, quantity int, ratings []Rating, sellerID string) error {
    exists, err := s.ProductExists(ctx, productID)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("the product %s does not exist", productID)
    }

    productJSON, err := ctx.GetStub().GetState(productID)
    if err != nil {
        return fmt.Errorf("failed to read product %s from world state: %v", productID, err)
    }
    if productJSON == nil {
        return fmt.Errorf("product %s does not exist", productID)
    }

    var existingProduct Product
    err = json.Unmarshal(productJSON, &existingProduct)
    if err != nil {
        return fmt.Errorf("failed to unmarshal existing product JSON: %v", err)
    }

    updatedProduct := Product{
        ProductID: productID,
        Name:      name,
        Price:     price,
        Quantity:  quantity,
        Ratings:   ratings,
        SellerID:  sellerID,
        Sold:      existingProduct.Sold,
    }

    updatedProductJSON, err := json.Marshal(updatedProduct)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(productID, updatedProductJSON)
}





func (s *SmartContract) DeleteProduct(ctx contractapi.TransactionContextInterface, productID string) error {
  exists, err := s.ProductExists(ctx, productID)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the product %s does not exist", productID)
  }

  return ctx.GetStub().DelState(productID)
}




func (s *SmartContract) ProductExists(ctx contractapi.TransactionContextInterface, productID string) (bool, error) {
  productJSON, err := ctx.GetStub().GetState(productID)
  if err != nil {
    return false, fmt.Errorf("failed to read from world state: %v", err)
  }

  return productJSON != nil, nil
}




func (s *SmartContract) GetAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
  resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
  if err != nil {
    return nil, err
  }
  defer resultsIterator.Close()

  var products []*Product
  for resultsIterator.HasNext() {
    queryResponse, err := resultsIterator.Next()
    if err != nil {
      return nil, err
    }

    var product Product
    err = json.Unmarshal(queryResponse.Value, &product)
    if err != nil {
      return nil, err
    }
    products = append(products, &product)
  }

  return products, nil
}



func (s *SmartContract) GetProductAverageRating(ctx contractapi.TransactionContextInterface, productID string) (float64, error) {
    productJSON, err := ctx.GetStub().GetState(productID)
    if err != nil {
        return 0, fmt.Errorf("failed to read product %s from world state: %v", productID, err)
    }
    if productJSON == nil {
        return 0, fmt.Errorf("product %s does not exist", productID)
    }

    var product Product
    err = json.Unmarshal(productJSON, &product)
    if err != nil {
        return 0, fmt.Errorf("failed to unmarshal product JSON: %v", err)
    }

    if len(product.Ratings) == 0 {
        return 0, nil // no ratings yet
    }

    var total int
    for _, rating := range product.Ratings {
        total += rating.Score
    }

    avg := float64(total) / float64(len(product.Ratings))
    return avg, nil
}



func (s *SmartContract) GetProductsBySeller(ctx contractapi.TransactionContextInterface, sellerID string) ([]*Product, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, fmt.Errorf("failed to get state by range: %v", err)
    }
    defer resultsIterator.Close()

    var products []*Product
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, fmt.Errorf("failed to iterate results: %v", err)
        }

        var product Product
        err = json.Unmarshal(queryResponse.Value, &product)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal product: %v", err)
        }

        if product.SellerID == sellerID {
            products = append(products, &product)
        }
    }

    return products, nil
}



func (s *SmartContract) PurchaseProduct(ctx contractapi.TransactionContextInterface, productID string, purchaseQuantity int) error {
    productJSON, err := ctx.GetStub().GetState(productID)
    if err != nil {
        return fmt.Errorf("failed to read product %s from world state: %v", productID, err)
    }
    if productJSON == nil {
        return fmt.Errorf("product %s does not exist", productID)
    }

    var product Product
    err = json.Unmarshal(productJSON, &product)
    if err != nil {
        return fmt.Errorf("failed to unmarshal product JSON: %v", err)
    }

    if purchaseQuantity <= 0 {
        return fmt.Errorf("purchase quantity must be positive")
    }

    if product.Quantity < purchaseQuantity {
        return fmt.Errorf("not enough quantity available: requested %d, available %d", purchaseQuantity, product.Quantity)
    }

    product.Quantity -= purchaseQuantity
    product.Sold += purchaseQuantity

    updatedProductJSON, err := json.Marshal(product)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(productID, updatedProductJSON)
}



func (s *SmartContract) AddRatingToProduct(ctx contractapi.TransactionContextInterface, productID string, ratingID string, score int, comment string) error {
    productJSON, err := ctx.GetStub().GetState(productID)
    if err != nil {
        return fmt.Errorf("failed to read product %s from world state: %v", productID, err)
    }
    if productJSON == nil {
        return fmt.Errorf("product %s does not exist", productID)
    }

    var product Product
    err = json.Unmarshal(productJSON, &product)
    if err != nil {
        return fmt.Errorf("failed to unmarshal product JSON: %v", err)
    }

    newRating := Rating{
	    Comment:   comment,
        ID:        ratingID,
        ProductID: productID,
        Score:     score,
    }

    product.Ratings = append(product.Ratings, newRating)

    updatedProductJSON, err := json.Marshal(product)
    if err != nil {
        return fmt.Errorf("failed to marshal updated product JSON: %v", err)
    }

    return ctx.GetStub().PutState(productID, updatedProductJSON)
}



func (s *SmartContract) GetProductRatings(ctx contractapi.TransactionContextInterface, productID string) ([]Rating, error) {
    productJSON, err := ctx.GetStub().GetState(productID)
    if err != nil {
        return nil, fmt.Errorf("failed to read product %s from world state: %v", productID, err)
    }
    if productJSON == nil {
        return nil, fmt.Errorf("product %s does not exist", productID)
    }

    var product Product
    err = json.Unmarshal(productJSON, &product)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal product JSON: %v", err)
    }

    return product.Ratings, nil
}



func (s *SmartContract) SearchProductsByName(ctx contractapi.TransactionContextInterface, searchTerm string) ([]*Product, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var matchingProducts []*Product
    lowerSearchTerm := strings.ToLower(searchTerm)

    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var product Product
        err = json.Unmarshal(queryResponse.Value, &product)
        if err != nil {
            return nil, err
        }

        if strings.Contains(strings.ToLower(product.Name), lowerSearchTerm) {
            matchingProducts = append(matchingProducts, &product)
        }
    }

    return matchingProducts, nil
}



func main() {
  productChaincode, err := contractapi.NewChaincode(&SmartContract{})
  if err != nil {
    log.Panicf("Error creating product chaincode: %v", err)
  }

  if err := productChaincode.Start(); err != nil {
    log.Panicf("Error starting product chaincode: %v", err)
  }
}