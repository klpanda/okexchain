# ERC20 on OKChain
  
ERC20 contract source code references [here](./erc20/erc20.sol).

### Create contract

Create a contract with 4 parameters ```_initialSupply, _name, _decimals, _symbol```: 

```text
_initialSupply: 1000000000
_name: token
_decimals: 18
_symbol: TOKEN
```

```bash
okchaincli tx evm create --code_file=./erc20/erc20.bc \
--from=$(okchaincli keys show -a alice) \
--args="1000000000 token 18 TOKEN" \
--abi_file="./erc20/erc20.abi" \
--gas=10000000 \
--fees 1tokt
```

Creating a contract will consume a lot of gas. The above command specifies the gas amount to 10000000.

After the contract is created successfully, check the transaction information according to txHash, where the address corresponding to the new_contract part in events is the new contract address ```okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc```

```bash
okchaincli query tx AB74A7790A2F7B5ECA08D0FCFB748C22207394919379D8FE04415AAAC5BA7F85
```

response:

```json
...
    {
      "type": "new_contract",
      "attributes": [
        {
          "key": "address",
          "value": "okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc"
        }
      ]
    },
...
```

### Query contract code

```bash
okchaincli query evm code okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc
```

REST API

```bash
curl http://127.0.0.1:1317/evm/code/okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc
```

### Query sybmol

Call the ```symbol``` method of the contract to query the token of ERC20

```bash
okchaincli query evm call $(okchaincli keys show -a alice) \
okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc symbol ./erc20/erc20.abi
```

response：

```json
{"Gas":9627,"Result":["TOKEN"]}
```

### Query name

Call the ```name``` method of the contract to query the name of ERC20

```bash
okchaincli query evm call $(okchaincli keys show -a alice) \
okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc name ./erc20/erc20.abi
```

response：

```json
{"Gas":9584,"Result":["token"]}
```

### Query totalSupply

Call the ```totalSupply``` method of the contract to query the totalSupply of ERC20


```bash
okchaincli query evm call $(okchaincli keys show -a alice) \
okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc totalSupply ./erc20/erc20.abi
```

response：

```json
{"Gas":7404,"Result":[1000000000]}
```

### Query decimals

Call the ```decimals``` method of the contract to query the decimals of ERC20


```bash
okchaincli query evm call $(okchaincli keys show -a alice) \
okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc decimals ./erc20/erc20.abi
```

response：

```json
{"Gas":7407,"Result":[18]}
```

### Query balances
  
The balance is initially set to `msg.sender`, which is the account that created the contract.

Query balance of alice by calling the ```balanceOf``` method of the contract：

```bash
okchaincli q evm call $(okchaincli keys show -a alice) \
okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc balanceOf ./erc20/erc20.abi \
--args "$(okchaincli keys show -a alice)"
```

response：

```json
{"Gas":7572,"Result":[1000000000]}
```

### Transfer


Transfer 10000 TOKEN to ```okchain1g4lvreq7c20sq7p6nphsp9qw29a2a3q40favg2``` by calling the ```transfer``` method of the contract:

```bash
okchaincli tx evm call --from=$(okchaincli keys show -a alice) \
--contract_addr=okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc \
--method=transfer \
--abi_file=./erc20/erc20.abi \
--args="okchain1g4lvreq7c20sq7p6nphsp9qw29a2a3q40favg2 10000" \
--gas=100000 --fees=1tokt
```
  
* query balance of ```okchain1g4lvreq7c20sq7p6nphsp9qw29a2a3q40favg2``` in contract

```bash
okchaincli query evm call $(okchaincli keys show -a alice) okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc \
balanceOf ./erc20/erc20.abi --args="okchain1g4lvreq7c20sq7p6nphsp9qw29a2a3q40favg2"
```

response：

```json
{"Gas":7572,"Result":[10000]}
```

* query balance of alice in contract

```bash
okchaincli query evm call $(okchaincli keys show -a alice) okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc \
balanceOf ./erc20/erc20.abi  --args=$(okchaincli keys show -a alice)
```

response：

```json
{"Gas":7572,"Result":[999990000]}
```

* query the event logs

```bash
okchaincli q evm logs EC40577049704BFA2ABBC90067FFCF9A4F3D1985FABA4077F0DF6081598A8206
```

response：

```json
{
  "logs": [
    {
      "address": "okchain1eq237struxpudlvtmr56f94zjv5uw0jm4r6ehc",
      "topics": [
        "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
        "000000000000000000000000d6e1bb6fa095f0e362054a5cd7e4231f5c0d2cd9",
        "000000000000000000000000457ec1e41ec29f00783a986f00940e517aaec415"
      ],
      "data": "0000000000000000000000000000000000000000000000000000000000002710",
      "blockNumber": "6030499",
      "transactionHash": "ec40577049704bfa2abbc90067ffcf9a4f3d1985faba4077f0df6081598a8206",
      "transactionIndex": "0",
      "blockHash": "0000000000000000000000000000000000000000000000000000000000000000",
      "logIndex": "0",
      "removed": false
    }
  ]
}
```