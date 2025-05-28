package common

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dumie-org/dumie-cli/awsutils"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func generateKeyPairName() string {
	randomString := uuid.New().String()[:23]
	return fmt.Sprintf("dumie-key-pair-%s", randomString)
}

func GenerateKeyPair(client *ec2.Client) (string, error) {
	existingKeyName, err := GetKeyPairName()
	if err != nil {
		return "", fmt.Errorf("failed to get key pair name: %v", err)
	}

	// Use existing key pair from config
	if existingKeyName != "" {
		return existingKeyName, nil
	}

	newKeyName := generateKeyPairName()
	privateKeyPath := filepath.Join(".", fmt.Sprintf("%s.pem", newKeyName))

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	createKeyPairInput := &ec2.CreateKeyPairInput{
		KeyName: aws.String(newKeyName),
		KeyType: types.KeyTypeRsa,
	}

	result, err := client.CreateKeyPair(context.TODO(), createKeyPairInput)
	if err != nil {
		return "", fmt.Errorf("failed to create key pair in AWS: %v", err)
	}

	privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create private key file: %v", err)
	}
	defer privateKeyFile.Close()

	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return "", fmt.Errorf("failed to write private key to file: %v", err)
	}

	return *result.KeyName, nil
}

func GetKeyPairName() (string, error) {
	config, err := awsutils.LoadAWSConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %v (hint: run `dumie configure` first)", err)
	}
	return config.KeyPairName, nil
}
