# Quark
This repository contains the source code for the quark contract language.

## Syntax
An example 'SellItem' specification is defined below:
```
spec SellItem dID to cID
  create Offer [item,price,quality]
  detach Pay [amount,address,shippingtype,deadline=5]
  discharge Delivery [deadline=5]
```
> **Note:** Every contract specification **MUST** include a deadline value in the detach and discharge event argument lists.

```
spec SPEC_NAME DEBTOR to CREDITOR
  create EVENT [ARG_LIST]
  detach EVENT [ARG_LIST, deadline=X]
  discharge EVENT [ARG_LIST, deadline=Y]
```
Where;

```
SPEC_NAME: The name of the specification
```

```
| Element       | Description                | Examples             |
| ------------- | ---------------------------------- | -------------------- |
| SPEC_NAME     | The name of the specification.     | SellItem, Refund     |
| ------------- | ---------------------------------- |
| DEBTOR        | The ID/name of the person in debt. | dID, debtorName      |
|       | to the creditor.             |            |
| EVENT     | The name of the event.       | Offer, Pay, Delivery |  
| ARG_LIST    | The list of arguments associated   |            |
|       | with a given event.          |            |  
```