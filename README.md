# Rent Receipt Generator
A simple script to generate rent receipts
## Requirements
- Golang
- Chrome
## Installation
```bash
go get -u github.com/hysem/itr
```

## Config File
```yaml
landlord:
  name: #land owner name
  address: #address of the land owner
  pan: #pan card number if given will be printed below the land owner name
tenant:
  name: #tenant name
  address: # address of tenant (not used currently)
  pan: #pan of tenant (not used currently)
financial_year: 2021 #fy for which the receipt to be generated
rent: 8000 # rent amount

```

## Generating receipts
- Make a copy of [config.yml](https://github.com/hysem/itr/blob/master/config.yaml) (say urname.yml)
- Run the tool 
```bash
itr -config urname.yml