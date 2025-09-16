package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/nicolasvancan/monvandb/src/database"
	"github.com/nicolasvancan/monvandb/src/server"
	"github.com/nicolasvancan/monvandb/src/system"

	"encoding/json"

	utils "github.com/nicolasvancan/monvandb/src/utils"
)

type Configuration struct {
	Port          int              `json:"port"`
	TLS           TLSConfiguration `json:"tls"`
	AdminUser     string           `json:"admin_user"`
	AdminPassword string           `json:"admin_password"`
}

type TLSConfiguration struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

func LoadConfiguration() (*Configuration, error) {
	// Load configuration from config.json file
	admin_user := os.Getenv("MONVANDB_ADMIN_USER")
	admin_password := os.Getenv("MONVANDB_ADMIN_PASSWORD")

	if admin_user == "" && admin_password == "" {
		admin_user = "admin"
		admin_password = "password"
	}

	if _, err := os.OpenFile(utils.GetPath("config"), os.O_RDONLY, 0644); err != nil {
		// If the file does not exist, create a default one
		// Return the configuration

		return &Configuration{
			Port: 8180,
			TLS: TLSConfiguration{
				Enabled:  true,
				CertFile: utils.GetPath("base") + string(os.PathSeparator) + "server.crt",
				KeyFile:  utils.GetPath("base") + string(os.PathSeparator) + "server.key",
			},
			AdminUser:     admin_user,
			AdminPassword: admin_password,
		}, nil
	}

	// If the file exists, read it
	data, err := os.ReadFile(utils.GetPath("config"))
	if err != nil {
		return nil, err
	}

	var config Configuration
	err = json.Unmarshal(data, &config)

	if err != nil {
		return nil, err
	}

	// fillup empty fields
	if config.Port == 0 {
		config.Port = 8180
	}
	if config.TLS.CertFile == "" {
		config.TLS.CertFile = utils.GetPath("base") + string(os.PathSeparator) + "server.crt"
	}
	config.TLS.Enabled = true // default is true
	if config.TLS.KeyFile == "" {
		config.TLS.KeyFile = utils.GetPath("base") + string(os.PathSeparator) + "server.key"
	}
	if config.AdminUser == "" {
		config.AdminUser = admin_user
	}
	if config.AdminPassword == "" {
		config.AdminPassword = admin_password
	}

	return &config, nil
}

// Initialize all database connections, caches, etc.
// If there is no certificate, create a self-signed one

func setup() (*Configuration, error) {
	config, err := LoadConfiguration()

	if err != nil {
		return nil, err
	}

	fmt.Println("Configuration loaded:", config)

	// Create base folders
	create_base_databases(config)

	// Generate self-signed certificate if TLS is enabled and cert files do not exist
	if config.TLS.Enabled && !utils.FileExists(config.TLS.CertFile) && !utils.FileExists(config.TLS.KeyFile) {
		fmt.Println("Generating self-signed certificate...")
		err := generateSelfSignedCert(config.TLS.CertFile, config.TLS.KeyFile)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func generateSelfSignedCert(certFile, keyFile string) error {
	// Generate a private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"MonvanDB"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost", "monvandb"},
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Save certificate to file
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
	if err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Save private key to file
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	err = pem.Encode(keyOut, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyDER,
	})
	if err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	fmt.Printf("Generated self-signed certificate: %s\n", certFile)
	fmt.Printf("Generated private key: %s\n", keyFile)
	return nil
}

func create_base_databases(config *Configuration) {
	// Create system database if not exists

	db, err := database.CreateDatabase("system")

	if err != nil {
		panic(err)
	}

	// Create system tables
	db.CreateTable("users", []database.Column{
		{Name: "id", Type: database.COL_TYPE_STRING, Primary: true, Nullable: false},
		{Name: "password", Type: database.COL_TYPE_STRING, Nullable: false},
		{Name: "roles", Type: database.COL_TYPE_BLOB, Nullable: true},
	}, false, false)

	db.CreateTable("roles", []database.Column{
		{Name: "id", Type: database.COL_TYPE_STRING, Primary: true, Nullable: false},
		{Name: "admin", Type: database.COL_TYPE_BOOL, Nullable: false},
		{Name: "permissions", Type: database.COL_TYPE_BLOB, Nullable: true},
	}, false, false)

	// Create base roles and users
	system.CreateRole("admin", map[string]interface{}{"admin": true})
	system.CreateUser(config.AdminUser, config.AdminPassword, []string{"admin"})
	system.LoadAllRoles()
	err = system.LoadAllUsersInMemory()
	if err != nil {
		panic(err)
	}

	user, err := system.GetUser(config.AdminUser)
	if err != nil {
		panic(err)
	}

	fmt.Println("user loaded:", user)
	fmt.Println("Admin user created with username:", config.AdminUser, "and password:", config.AdminPassword)
}

func main() {
	config, err := setup()
	fmt.Println(config)
	if err != nil {
		fmt.Println("Error setting up:", err)
		return
	}

	server := server.NewServer(config.Port)
	fmt.Println(os.Getenv("MONVANDB_ADMIN_USER"))

	server.Run(config.TLS.CertFile, config.TLS.KeyFile)
	// Example usage of the parser
	// Uncomment the following lines to test the SQL parser
	// parsed, err := sqlparser.Parse("SELECT a + 5 from tab1")
	// if err != nil {
	// 	fmt.Println("Error parsing SQL:", err)
	// 	return
	// }

	// fmt.Printf("Parsed SQL: %T\n", parsed.(*sqlparser.Select).SelectExprs[0].(*sqlparser.AliasedExpr).Expr.(*sqlparser.BinaryExpr))
	// fmt.Printf("Parsed SQL: %v\n", parsed)
	// parser, pass, err := monvan_parser.Parse("SHOW TABLES system")
	// fmt.Println(pass)
	// if err != nil {
	// 	fmt.Println("Error parsing SQL:", err)
	// 	return
	// }
	// // print type of variable parser
	// fmt.Printf("Parsed SQL: %T\n", parser)
	// fmt.Printf("Parsed SQL: %v\n", parser)

}
