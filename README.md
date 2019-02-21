## Tx Lifecycle Test in Geth

### Instructions to run 
```
$ mkdir -p go/src/panghalamit 

$ cd go/src/panghalamit

$ git clone git@github.com:panghalamit/tx-lifecycle-test.git

$ git clone git@github.com:panghalamit/go-ethereum
```

#### To build modified go-ethereum
```
$ cd go-ethereum

$ make
```
#### To build test client
```
$ cd tx-lifecycle-test

$ go build .
```

#### Private Network Setup 

Follow the instructions given [here](https://hackernoon.com/setup-your-own-private-proof-of-authority-ethereum-network-with-geth-9a0a3750cda8) to set up private network. We will use test client to interact with the network.

##### Gotchas
* Make sure config values you use in genesis config are same as ones you pass while running the node. (Network Id should be same as chain id in genesis.json)

* You will use test accounts, So make sure you allocate funds to them in your genesis config file. 

* Keep the test accounts file in a directory, test accounts directory will be passed as argument to the client.

* Account files will have names like 'UTC--2019-02-20T03-04-56.953368012Z--eb84ed27085727da77740121cf80db89750662a3'

* There should atleast be 3 account files for all test scenarios to work.


#### Running Test

To run the client

```
$ cd tx-lifecycle-test

$ ./tx-lifecycle-test <rpcAddrOfOneOfTheNodes> <pathToAccountsDir>
```

Lets say rpc addr is http://localhost:8501 .
and test accounts are in test-accounts in tx-lifecycle-test . To run the test we do
```
$ ./tx-lifecycle-test http://localhost:8501 test-accounts

```

### References

* [Go Ethereum](https://github.com/ethereum/go-ethereum)
* [Go Ethereum Book](https://goethereumbook.org/en/)
* [Hackernoon post on setting up private node](https://hackernoon.com/setup-your-own-private-proof-of-authority-ethereum-network-with-geth-9a0a3750cda8)

