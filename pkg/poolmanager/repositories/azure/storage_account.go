package azure

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/storage"
)

var theAccountInstance *StorageAccount

// StorageAccount in Azure
type StorageAccount struct {
	accountName string
	accountKey  string
	client      storage.Client
}

// SetStorageAccount sets name and key to use
func SetStorageAccount(accountName, accountKey string) error {
	theAccountInstance = &StorageAccount{
		accountName: accountName,
		accountKey:  accountKey,
	}

	client, err := storage.NewBasicClient(accountName, accountKey)
	if err != nil {
		log.Println(err)
		return err
	}

	theAccountInstance.client = client
	return nil
}

// GetStorageAccountInstance gets the instance
func GetStorageAccountInstance() *StorageAccount {
	return theAccountInstance
}

// GetClient gets storage account client
func (account *StorageAccount) GetClient() storage.Client {
	return account.client
}
