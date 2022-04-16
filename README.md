# OhNo Vulnerability Scanner

OhNo is a vulnerability scanner for Python scripts that stands on [Bandit](https://github.com/PyCQA/bandit).

## Usage

### New Scan
#### Request
```bash
curl -X POST localhost:8080/newscan -d '{ "url": "https://github.com/yinkar/PyPorte" }'
```

#### Response
```json
{
    "scan_id": "cc5b491f-b6e3-48f0-acc9-1695bf70ef6d"
}
```

### View Scan Results
#### Request
```bash
curl localhost:8080/scan/cc5b491f-b6e3-48f0-acc9-1695bf70ef6d
```

#### Response
```json
{"safety":true,"results":[{"scan_id":"cc5b491f-b6e3-48f0-acc9-1695bf70ef6d","code":"54 \telse:\n55 \t\tout = input('Type filename of the json file: ')\n56 \n","filename":"/code/PyPorte.py","issue_severity":"HIGH","created_at":"2022-04-16 11:33:22.9944275+03:00"}]}
```

