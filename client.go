package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/panghalamit/go-ethereum/accounts"
	"github.com/panghalamit/go-ethereum/accounts/keystore"
	"github.com/panghalamit/go-ethereum/core/types"
	"github.com/panghalamit/go-ethereum/ethclient"
)

type TestConfig struct {
	PathToAccountsDir string
	Scenario1         Scenario1Config
	Scenario2         Scenario2Config
	Scenario3         Scenario3Config
}

type Scenario1Config struct {
	Value    json.Number
	GasLimit uint64
}

type Scenario2Config struct {
	Value           json.Number
	GasLimit        uint64
	NumTxs          int
	RescueTxTimeout int
}

type Scenario3Config struct {
	Value                       json.Number
	GasLimit                    uint64
	MinExecutablesPerAccount    uint32
	MaxExecutablesTotal         uint32
	MaxNonExecutablesPerAccount uint32
	MaxNonExecutablesTotal      uint32
	NumTxs                      int
}

var config TestConfig

// InitTestConfig inititializes config parameters from json file
func InitTestConfig() (*TestConfig, error) {
	jsonBytes, err := ioutil.ReadFile("config-test.json")
	if err != nil {
		return &config, err
	}
	err1 := json.Unmarshal(jsonBytes, &config)
	if err1 != nil {
		return &config, err1
	}
	return &config, nil
}

// ImportKsAccount imports accounts into keystore given path to account's file
func ImportKsAccount(ks *keystore.KeyStore, pathToAccount string) (*accounts.Account, error) {
	jsonBytes, err := ioutil.ReadFile(pathToAccount)
	if err != nil {
		return &accounts.Account{}, err
	}
	pwd := ""
	account, err := ks.Import(jsonBytes, pwd, pwd)
	if err != nil {
		return &accounts.Account{}, err
	}
	return &account, nil
}

// Init creates a keystore and imports test accounts from list of account files
func Init(accfiles []string, ks *keystore.KeyStore) ([]*accounts.Account, error) {
	acclist := make([]*accounts.Account, len(accfiles))
	var err error
	for ind, acc := range accfiles {
		acclist[ind], err = ImportKsAccount(ks, acc)
		if err != nil {
			return nil, err
		}
	}
	return acclist, nil
}

// FilePathWalkDir returns all files present in a directory
func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// CreateTransaction creates new transaction with given sender,receiver, nonce, value, gasPrice, chainId
func CreateTransaction(ks *keystore.KeyStore, sender *accounts.Account, receiver *accounts.Account, passphrase string, nonce uint64, value *big.Int, data []byte, gasPrice *big.Int, gasLimit uint64, chainId *big.Int) (*types.Transaction, error) {
	tx := types.NewTransaction(nonce, receiver.Address, value, gasLimit, gasPrice, data)
	signedTx, err := ks.SignTxWithPassphrase(*sender, passphrase, tx, chainId)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

// Scenario1 sends a single transaction from a account
func Scenario1(cl *ethclient.Client, acclist []*accounts.Account, ks *keystore.KeyStore) {
	fmt.Printf("\n -------------------------Scenario1 : Sending a single transfer transaction-------------------------\n")
	sender := acclist[0]
	receiver := acclist[1]
	nonce, err1 := cl.PendingNonceAt(context.Background(), sender.Address)
	if err1 != nil {
		fmt.Printf("err1\n")
		log.Fatal(err1)
	}

	balance, err2 := cl.BalanceAt(context.Background(), sender.Address, nil)
	if err2 != nil {
		log.Fatal(err2)
	}
	fmt.Printf("Balance of account %s : %v \n", sender.Address.Hex(), balance)

	fmt.Printf("Nonce value for account %s : %v\n", sender.Address.Hex(), nonce)
	value, ok := new(big.Int).SetString(config.Scenario1.Value.String(), 10)
	if !ok {
		value = big.NewInt(1000000000000000000)
	}
	gasLimit := config.Scenario1.GasLimit
	gasPrice, err3 := cl.SuggestGasPrice(context.Background())
	if err3 != nil {
		fmt.Printf("err3\n")
		log.Fatal(err3)
	}
	fmt.Printf("Suggested gas price %v\n", gasPrice)
	chainID, err4 := cl.NetworkID(context.Background())
	if err4 != nil {
		fmt.Printf("err4\n")
		log.Fatal(err4)
	}
	fmt.Printf("ChainId = %v\n", chainID)
	var data []byte
	signedTx, err5 := CreateTransaction(ks, sender, receiver, "", nonce, value, data, gasPrice, gasLimit, chainID)
	if err5 != nil {
		fmt.Printf("err5\n")
		log.Fatal(err5)
	}
	fmt.Printf("Signed transaction successfully :%v \n", signedTx.Hash().Hex())
	err6 := cl.SendTransaction(context.Background(), signedTx)
	if err6 != nil {
		fmt.Printf("err6\n")
		log.Fatal(err6)
	}
	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
}

// Scenario2 sends multiple transactions from a accounts with random nonce gap
func Scenario2(cl *ethclient.Client, acclist []*accounts.Account, ks *keystore.KeyStore) {
	fmt.Printf("\n-------------------------Scenario2 : Sending multiple transfer transactions with gap in nonce values-------------------------\n")
	sender := acclist[2]
	receiver := acclist[0]
	rand.Seed(time.Now().UTC().UnixNano())
	nonce, err1 := cl.PendingNonceAt(context.Background(), sender.Address)
	if err1 != nil {
		log.Fatal(err1)
	}
	chainID, err2 := cl.NetworkID(context.Background())
	if err2 != nil {
		log.Fatal(err2)
	}
	value, ok := new(big.Int).SetString(config.Scenario2.Value.String(), 10)
	if !ok {
		value = big.NewInt(1000000000000000000)
	}
	gasLimit := config.Scenario2.GasLimit
	gasPrice, err3 := cl.SuggestGasPrice(context.Background())
	var data []byte
	if err3 != nil {
		log.Fatal(err3)
	}
	for i := 0; i < config.Scenario2.NumTxs; i++ {
		toss := rand.Float32()
		if toss < 0.3 {
			gasPrice = big.NewInt(2)
		} else if toss > 0.7 {
			gasPrice = big.NewInt(3)
		}
		fmt.Printf("Gas price %v\n", gasPrice)
		fmt.Printf("ChainId = %v\n", chainID)
		signedTx, err4 := CreateTransaction(ks, sender, receiver, "", nonce+uint64(i+1), value, data, gasPrice, gasLimit, chainID)
		if err4 != nil {
			log.Fatal(err4)
		}
		fmt.Printf("Signed transaction successfully :%v \n", signedTx.Hash().Hex())
		err5 := cl.SendTransaction(context.Background(), signedTx)

		if err5 != nil {
			fmt.Println(err5)
			//log.Fatal(err6)
		} else {
			fmt.Printf("tx sent with hash: %s, nonce: %v, value: %v\n", signedTx.Hash().Hex(), nonce+uint64(i+1), value)
		}
	}

	// get latest block
	currBlockHeader, err6 := cl.HeaderByNumber(context.Background(), nil)
	if err6 != nil {
		fmt.Println(err6)
		return
	}
	blockNumber := currBlockHeader.Number.Uint64()

	fmt.Printf("Using block height as timeout to send rescue Tx (Tx with gap value as nonce)\n")
	for curr := blockNumber; curr < blockNumber+uint64(config.Scenario2.RescueTxTimeout); {
		// get latest block
		currBlockHeader, err6 = cl.HeaderByNumber(context.Background(), nil)
		if err6 != nil {
			fmt.Println(err6)
			return
		}
		curr = currBlockHeader.Number.Uint64()
		fmt.Printf("Block height %v \n", curr)
		time.Sleep(15 * time.Second)
	}
	// wait for config.Scenario2.RescueTxTimeout block times before sending rescue tx
	signedRescueTx, err7 := CreateTransaction(ks, sender, receiver, "", nonce, value, data, gasPrice, gasLimit, chainID)
	if err7 != nil {
		log.Fatal(err7)
	}
	fmt.Printf("Signed transaction successfully :%v \n", signedRescueTx.Hash().Hex())
	err8 := cl.SendTransaction(context.Background(), signedRescueTx)
	if err8 != nil {
		log.Fatal(err8)
	} else {
		fmt.Printf("Rescue tx with pending nonce sent with hash: %s, nonce: %v, value: %v\n", signedRescueTx.Hash().Hex(), nonce, value)
	}

}

// Scenario3 sends multiple transactions from multiple accounts
func Scenario3(cl *ethclient.Client, acclist []*accounts.Account, ks *keystore.KeyStore) {
	fmt.Printf("\n-------------------------Scenario3 : Sending multiple transfer transactions from different accounts-------------------------\n")
	//set the same value in txpool config
	minExecutablesPerAccount := config.Scenario3.MinExecutablesPerAccount
	maxExecutablesTotal := config.Scenario3.MaxExecutablesTotal
	maxNonExecutablesPerAccount := config.Scenario3.MaxNonExecutablesPerAccount
	maxNonExecutablesTotal := config.Scenario3.MaxNonExecutablesTotal

	fmt.Printf("Testing for txpool config values as, accountslots : %v, globalslots : %v, accountqueue : %v, globalqueue : %v  \n", minExecutablesPerAccount, maxExecutablesTotal, maxNonExecutablesPerAccount, maxNonExecutablesTotal)
	gasPrice, err3 := cl.SuggestGasPrice(context.Background())
	if err3 != nil {
		log.Fatal(err3)
	}
	chainID, err4 := cl.NetworkID(context.Background())
	if err4 != nil {
		log.Fatal(err4)
	}
	fmt.Printf("ChainId = %v\n", chainID)

	value, ok := new(big.Int).SetString(config.Scenario3.Value.String(), 10)
	if !ok {
		value = big.NewInt(1000000000000000000)
	}
	gasLimit := config.Scenario3.GasLimit
	size := uint32(len(acclist))
	for i := uint32(0); i < uint32(config.Scenario3.NumTxs); i++ {
		sender := acclist[i%size]
		receiver := acclist[(i+1)%size]
		nonce, err1 := cl.PendingNonceAt(context.Background(), sender.Address)
		if err1 != nil {
			log.Fatal(err1)
		}
		gasPrice = big.NewInt(int64(i))
		fmt.Printf("Gas price %v\n", gasPrice)
		var data []byte
		signedTx, err5 := CreateTransaction(ks, sender, receiver, "", nonce, value, data, gasPrice, gasLimit, chainID)
		if err5 != nil {
			log.Fatal(err5)
		}
		fmt.Printf("Signed transaction successfully :%v \n", signedTx.Hash().Hex())
		err6 := cl.SendTransaction(context.Background(), signedTx)
		if err6 != nil {
			fmt.Println(err6)
			//log.Fatal(err6)
		}

	}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: ./TxLifeCycle <NodeRpcAddr> <ScenarioToRun>")
	}

	_, err3 := InitTestConfig()
	if err3 != nil {
		log.Fatal(err3)
	}
	fmt.Println("Successfully Read Test Configuration from config file")

	client, err := ethclient.Dial(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("We connected to a local node")
	accfiles, err1 := FilePathWalkDir(config.PathToAccountsDir)
	if err1 != nil {
		log.Fatal(err1)
	}
	if len(accfiles) < 3 {
		log.Fatal("Need atleast 3 accounts to test all scenarios")
	}

	scenarioToRun := os.Args[2]
	ks := keystore.NewKeyStore("./tmp", keystore.StandardScryptN, keystore.StandardScryptP)
	acclist, err2 := Init(accfiles, ks)
	if err2 != nil {
		log.Fatal(err2)
	}
	if scenarioToRun == "1" {
		Scenario1(client, acclist, ks)
	} else if scenarioToRun == "2" {
		Scenario2(client, acclist, ks)
	} else if scenarioToRun == "3" {
		Scenario3(client, acclist, ks)
	}
}
