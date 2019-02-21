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

### Test Scenarios

#### Scenario1

Sending a simple valid transaction to the client from one of the test account.
Transation sent is simple eth transfer transaction.  Rpc is used to communicate, here is set of functions transactions go through before mined.

At Node for which given transaction local, i.e., received directly via rpc call by client.

* Transaction is received at SendTx function in eth/api-backend.go 
* Transaction is sent to pool via addLocal(tx) in core/tx_pool.go
* internal addTx(tx,isLocal) is called. islocal determines the pricing constraints for tx
* internal add(tx, isLocal) is called, with a lock on pool object
  * return if tx is already in pool
* Basic validation (validateTx in core/tx_pool.go)  
  * check for size
  * negative value check
  * gasLimit check against block gasLimit (tx's gas can't be greater than block's limit)
  * signature validity check
  * gas price check of non-local tx
  * nonce check (can't be lower that current nonce of sender)
  * enough funds check
  * intrinsic gas check
* if basic validation passes and tx_pool is full, remove underpriced tx to make space for this.
* if tx replaces a pending tx (tx which can be processed given current state, i.e., given nonce is already in pending list), remove old, add this tx
* Notify subscribers that new tx is added to pending list
* tx isn't replacing a pending tx, enqueue to future queue using (enqueueTx in core/tx_pool.go)
  * add to queue if tx is new or better than older tx
* add to disk journal if tx is local
* Run promotion check for transactions from sender, if they can be moved to pending list

Miner Listens to new tx events
* on receiving a new txEvent at (miner/worker.go->mainLoop())
  * commit and update if not mining (miner/worker.go -> commitTransactions)
  * commit new work if mining (miner/worker.go -> commitNewWork)
    * fetches all pending tx from pool, separates into local and remote
    * first commits local txs and then remote (each ordered by tx fees and nonce)


#### Scenario2

Sending multiple transactions from a single account with gap between nonces.

* Txs will be added to future queue, won't be promoted unless transaction with gap nonces are sent.
* Txs with nonces too far in future will be rejected

#### Scenario3

Sending multiple transactions from different accounts. 

* Sending tx with different gas prices to see order in which transactions are mined
* To make sure all txs are not included in single block, play with block gasLimit or timeout which determines the wait a miner does for a new tx before committing to current snapshot for mining.


### References

* [Go Ethereum](https://github.com/ethereum/go-ethereum)
* [Go Ethereum Book](https://goethereumbook.org/en/)
* [Hackernoon post on setting up private node](https://hackernoon.com/setup-your-own-private-proof-of-authority-ethereum-network-with-geth-9a0a3750cda8)

