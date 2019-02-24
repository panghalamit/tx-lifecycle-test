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

$ ./tx-lifecycle-test <rpcAddrOfOneOfTheNodes> <pathToAccountsDir> <ScenarioToTest>
```

Lets say rpc addr is http://localhost:8501 .
and test accounts are in test-accounts in tx-lifecycle-test . To run the test for scenario1 we do
```
$ ./tx-lifecycle-test http://localhost:8501 test-accounts 1 

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
* Say current pending nonce for a account is x, send muliple Tx from account with consecutive nonces starting from x+1.
* Wait for 2-3 blocktime.
* All Txs will be in future queue.
* Now send a transaction with nonce x, All tx send in previous step will be promoted to pending and included in block eventually.

#### Scenario3

Sending multiple transactions from different accounts. 

* This testcase to see how txpool config parameters govenn txpool management.
* accountslots determines minimum allowed pending txs from a account.
* globalslots determines maximum allowed pending tx from all accounts to prevent DOS.
  * if total size of pending list is above globalslots value, non-local txs are moved to queue with an attempt to honour accountslots if possible.
* accountqueue is cap on no. of txs that can be put to future queue for a account and globalqueue is maximum cap overall
  * non-local tx are discarded if total queued txs from a single account are above accountqueue. If globalqueue limit is crossed, then non-local txs are removed based on priority
* Test involves sending lot of txs from different accounts, such that set limits of above parameters are crossed.
* make sure to have large enough block time to allow large number of txs sent before they are mined in block and removed from pool.
* logs of node other than the one to which txs are sent should be observed. Since above params only remote tx are evicted or demoted by the policy


### References

* [Go Ethereum](https://github.com/ethereum/go-ethereum)
* [Go Ethereum Book](https://goethereumbook.org/en/)
* [Hackernoon post on setting up private node](https://hackernoon.com/setup-your-own-private-proof-of-authority-ethereum-network-with-geth-9a0a3750cda8)

