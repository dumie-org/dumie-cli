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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	// KeyName is the name of the EC2 key pair used for instance access
	KeyName = "dumie-key-pair"
)

func GenerateKeyPair(client *ec2.Client) (string, error) {
	privateKeyPath := filepath.Join(".", fmt.Sprintf("%s.pem", KeyName))
	if _, err := os.Stat(privateKeyPath); err == nil {
		fmt.Printf("Key pair file already exists at %s, skipping generation\n", privateKeyPath)
		return KeyName, nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	createKeyPairInput := &ec2.CreateKeyPairInput{
		KeyName: aws.String(KeyName),
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
