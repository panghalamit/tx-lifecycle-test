package main

import (
	"context"
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

const ()

// ImportKsAccount imports accounts into keystore given path to account's file
func ImportKsAccount(ks *keystore.KeyStore, pathToAccount string) (*accounts.Account, error) {
	jsonBytes, err := ioutil.ReadFile(pathToAccount)
	if err != nil {
		return &accounts.Account{}, err
	}
	pwd := ""
	account, err := ks.Import(jsonBytes, pwd, pwd)
	if err != nil {
		log.Fatal(err)
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
	value := big.NewInt(1000000000000000000)
	gasLimit := uint64(21000)
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
	value := big.NewInt(1000000000000000000)
	gasLimit := uint64(21000)
	gasPrice, err3 := cl.SuggestGasPrice(context.Background())
	var data []byte
	if err3 != nil {
		log.Fatal(err3)
	}
	for i := 0; i < 10; i++ {
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
	/*
		// get latest block
		currBlockHeader, err6 := cl.HeaderByNumber(context.Background(), nil)
		if err6 != nil {
			fmt.Println(err6)
			return
		}
		blockNumber = currBlockHeader.Nonce.Uint64()

		fmt.Printf("Using block height as timeout to send rescue Tx (Tx with gap value as nonce)\n")
		for curr := blockNumber; curr < blockNumber+2; {
			// get latest block
			currBlockHeader, err6 = cl.HeaderByNumber(context.Background(), nil)
			if err6 != nil {
				fmt.Println(err6)
				return
			}
			curr = currBlockHeader.Nonce.Uint64()
			fmt.Printf("Block height %v \n", curr)
			time.Sleep(15 * time.Second)
		}*/
	// fix this, ugly: block time in seconds (block height solution not working)
	// wait for 3 block times before sending rescue tx
	time.Sleep(time.Second * 3 * 15)
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
	//set the same value in txpool.accountslots
	minExecutablesPerAccount := uint32(10)
	maxExecutablesTotal := uint32(10)

	rand.Seed(time.Now().UTC().UnixNano())
	value := big.NewInt(1000000000000000000)
	gasLimit := uint64(21000)
	gasPrice, err3 := cl.SuggestGasPrice(context.Background())
	if err3 != nil {
		log.Fatal(err3)
	}
	toss := rand.Float32()
	if toss < 0.3 {
		gasPrice = big.NewInt(2)
	} else if toss > 0.7 {
		gasPrice = big.NewInt(3)
	}
	countExecutable := make([]uint32, len(acclist))
	countTotalExecutables := uint32(0)

	for i := uint32(0); i < 4*maxExecutablesTotal; i++ {
		j := rand.Intn(len(acclist))
		sender := acclist[j]
		receiver := acclist[(j+1)%len(acclist)]
		nonce, err1 := cl.PendingNonceAt(context.Background(), sender.Address)
		if err1 != nil {
			log.Fatal(err1)
		}
		fmt.Printf("Gas price %v\n", gasPrice)
		chainID, err4 := cl.NetworkID(context.Background())
		if err4 != nil {
			log.Fatal(err4)
		}
		fmt.Printf("ChainId = %v\n", chainID)
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
		} else {
			countExecutable[j]++
			countTotalExecutables++
			fmt.Printf("tx sent with hash: %s, nonce: %v, value: %v\n", signedTx.Hash().Hex(), nonce, value)
			if countTotalExecutables > maxExecutablesTotal {
				for ind, val := range countExecutable {
					if val > minExecutablesPerAccount {
						fmt.Printf("Txs from Account %v might get demoted\n", acclist[ind].Address.Hex())
					}
				}
			}
		}

	}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: ./TxLifeCycle <NodeRpcAddr> <pathToTestAccountsDirectory>")
	}

	client, err := ethclient.Dial(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("We connected to a local node")

	accfiles, err1 := FilePathWalkDir(os.Args[2])
	if err1 != nil {
		log.Fatal(err1)
	}
	if len(accfiles) < 3 {
		log.Fatal("Need atleast 3 accounts to test all scenarios")
	}
	ks := keystore.NewKeyStore("./tmp", keystore.StandardScryptN, keystore.StandardScryptP)
	acclist, err2 := Init(accfiles, ks)
	if err2 != nil {
		log.Fatal(err2)
	}
	Scenario1(client, acclist, ks)
	Scenario2(client, acclist, ks)
	Scenario3(client, acclist, ks)
}
