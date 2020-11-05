package azure

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories"
)

// DefinitionRepository azure blob storage
type definitionRepository struct {
	blobContainerName string
	blobClient        storage.BlobStorageClient
}

// NewAzureDefinitionRepository new Instance
func NewAzureDefinitionRepository(blobContainerName string) (repositories.DefinitionRepository, error) {
	account := GetStorageAccountInstance()
	blobClient := account.GetClient().GetBlobService()
	_, err := blobClient.GetContainerReference(blobContainerName).CreateIfNotExists(&storage.CreateContainerOptions{
		Access: storage.ContainerAccessTypePrivate,
	})

	if err != nil {
		return nil, err
	}

	return &definitionRepository{
		blobContainerName: blobContainerName,
		blobClient:        blobClient,
	}, nil
}

// Add implements repositories.DefinitionRepository
func (repo *definitionRepository) Add(deploymentID string, yaml []byte, parameters []byte) error {
	yamlBlob := fmt.Sprintf("%s-yaml.yaml", deploymentID)
	paramBlob := fmt.Sprintf("%s-param.yaml", deploymentID)

	cntr := repo.blobClient.GetContainerReference(repo.blobContainerName)
	blob := cntr.GetBlobReference(yamlBlob)
	if err := blob.CreateBlockBlobFromReader(bytes.NewReader(yaml), nil); err != nil {
		return err
	}

	blob = cntr.GetBlobReference(paramBlob)
	if err := blob.CreateBlockBlobFromReader(bytes.NewReader(parameters), nil); err != nil {
		return err
	}

	return nil
}

// Get implements repositories.DefinitionRepository
func (repo *definitionRepository) Get(deploymentID string) ([]byte, []byte, error) {
	yamlBlob := fmt.Sprintf("%s-yaml.yaml", deploymentID)
	paramBlob := fmt.Sprintf("%s-param.yaml", deploymentID)

	cntr := repo.blobClient.GetContainerReference(repo.blobContainerName)

	blob := cntr.GetBlobReference(yamlBlob)

	resp, err := blob.Get(nil)
	if err != nil {
		return nil, nil, err
	}

	yamlData, err := ioutil.ReadAll(resp)
	if err != nil {
		return nil, nil, err
	}

	blob = cntr.GetBlobReference(paramBlob)
	resp, err = blob.Get(nil)
	if err != nil {
		return nil, nil, err
	}

	paramData, err := ioutil.ReadAll(resp)
	if err != nil {
		return nil, nil, err
	}

	return yamlData, paramData, nil
}

// Delete implements repositories.DefinitionRepository
func (repo *definitionRepository) Delete(deploymentID string) error {
	yamlBlob := fmt.Sprintf("%s-yaml.yaml", deploymentID)
	paramBlob := fmt.Sprintf("%s-param.yaml", deploymentID)

	cntr := repo.blobClient.GetContainerReference(repo.blobContainerName)

	blob := cntr.GetBlobReference(yamlBlob)

	if _, err := blob.DeleteIfExists(nil); err != nil {
		return err
	}

	blob = cntr.GetBlobReference(paramBlob)

	if _, err := blob.DeleteIfExists(nil); err != nil {
		return err
	}
	return nil
}
