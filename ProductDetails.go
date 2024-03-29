package main;

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"encoding/json"
)

type ProductDetailsContract struct {
	contractapi.Contract
}

/**
*@dev Product() represents the product details
*/

type Product struct {
	ID              uint64 `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ManufactureDate uint64 `json:"manufactureDate"`
	BatchNumber     string `json:"batchNumber"`
	State ProductState `json:"state"`
}

/**
*@dev ProductHistory() represents the history of a product
*/

type ProductHistory struct {
	Timestamp uint64        `json:"timestamp"`
	Action    string        `json:"action"`
	Location  string        `json:"location"`
	State     ProductState `json:"state"`
}

/**
*@dev ProductState() represents the state of a product
*/

type ProductState int

const (
	PRODUCT_REGISTERED ProductState = iota
	QUALITY_ASSURANCE
	PRODUCT_TRANSIT
	PRODUCT_IN_INVENTORY
	PRODUCT_SOLD
	PRODUCT_RECALLED
	CONSUMPTION
	PENDING
	VALIDATING
	PUBLISHING
)

/**
@dev Init() initializes the chaincode
*/

func (c *ProductDetailsContract) Init(ctx contractapi.TransactionContextInterface) error {
	// Initialization later
	return nil
}

/**
*@dev AddProduct() adds a new product
*/

func (c *ProductDetailsContract) AddProduct(ctx contractapi.TransactionContextInterface, name string, description string, manufacturedDate uint64, batchNumber string) error {
	nextProductID, err := c.generateNextProductID(ctx)
	if err != nil {
		return err
	}

	product := Product{
		ID:              nextProductID,
		Name:            name,
		Description:     description,
		ManufactureDate: manufacturedDate,
		BatchNumber:     batchNumber,
	}

	err = ctx.GetStub().PutState(fmt.Sprintf("PRODUCT-%d", nextProductID), []byte(product));
	if err != nil {
		return fmt.Errorf("failed to put product on the ledger: %v", err)
	}

	return nil

}

/**
*@dev RetrieveProductDetails() retrieves the details of a product
*/

func (c *ProductDetailsContract) RetrieveProductDetails(ctx contractapi.TransactionContextInterface, productID uint64) (*Product, error) {
	productBytes, err := ctx.GetStub().GetState(fmt.Sprintf("PRODUCT-%d", productID))
	if err != nil {
		return nil, fmt.Errorf("failed to read product from the ledger: %v", err)
	}
	if productBytes == nil {
		return nil, fmt.Errorf("product with ID %d does not exist", productID)
	}

	product := new(Product)
	err = json.Unmarshal(productBytes, product)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product JSON: %v", err)
	}

	return product, nil
}

/**
*@dev UpdateProductState() updates the state of a product
*/

func (c *ProductDetailsContract) UpdateProductState(ctx contractapi.TransactionContextInterface, productID uint64, currentState ProductState) error {
	product, err := c.RetrieveProductDetails(ctx, productID)
	if err != nil {
		return err
	}

	/**
	*@dev check for valid state transitions
    */

	if product.State == PRODUCT_REGISTERED && currentState != PRODUCT_TRANSIT {
		return fmt.Errorf("invalid state transition")
	}

	product.State = currentState
	productBytes, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product JSON: %v", err)
	}

	err = ctx.GetStub().PutState(fmt.Sprintf("PRODUCT-%d", productID), productBytes)
	if err != nil {
		return fmt.Errorf("failed to put updated product state on the ledger: %v", err)
	}

	return nil
}

/**
*@dev LogProductMovement logs the movement of a product
*/

func (c *ProductDetailsContract) LogProductMovement(ctx contractapi.TransactionContextInterface, productID uint64, newLocation string) error {
	product, err := c.RetrieveProductDetails(ctx, productID)
	if err != nil {
		return err
	}

	productHistory := ProductHistory{
		Timestamp: uint64(ctx.GetStub().GetTxTimestamp().GetSeconds()),
		Action:    "Movement",
		Location:  newLocation,
		State:     product.State,
	}

	timestamp, _ := ctx.GetStub().GetTxTimestamp() // Error handling is not required here
    productHistory.Timestamp = uint64(timestamp.GetSeconds())

	historyKey := fmt.Sprintf("PRODUCT-%d-HISTORY", productID)
	existingHistoryBytes, err := ctx.GetStub().GetState(historyKey)
	if err != nil {
		return fmt.Errorf("failed to read product history from the ledger: %v", err)
	}

	var productHistories []ProductHistory
	if existingHistoryBytes != nil {
		err = json.Unmarshal(existingHistoryBytes, &productHistories)
		if err != nil {
			return fmt.Errorf("failed to unmarshal product history JSON: %v", err);
}

		// Unmarshal existing product histories
		err = json.Unmarshal(existingHistoryBytes, &productHistories)
		if err != nil {
			return fmt.Errorf("failed to unmarshal product history JSON: %v", err)
		}
	}

	// Append the new product history
	productHistories = append(productHistories, productHistory)

	// Marshal the updated product history
	updatedHistoryBytes, err := json.Marshal(productHistories)
	if err != nil {
		return fmt.Errorf("failed to marshal updated product history JSON: %v", err)
	}

	// Store the updated history on the ledger
	err = ctx.GetStub().PutState(historyKey, updatedHistoryBytes)
	if err != nil {
		return fmt.Errorf("failed to put updated product history on the ledger: %v", err)
	}

	return nil
}


