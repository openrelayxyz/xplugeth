import json

def test_cardinal():
    with open('./resources/control_card_data.json', 'r') as cf:
        control_card = json.load(cf)

    with open('./resources/test_card_data.json', 'r') as tf:
        test_card = json.load(tf)
        
    with open('./resources/control_plugeth_data.json', 'r') as cf:
        control_plugeth = json.load(cf)

    with open('./resources/test_plugeth_data.json', 'r') as tf:
        test_plugeth = json.load(tf)

    errors = []

    for i, item in enumerate(control_card):
        for k, v in item['result']['batch'].items():
            if type(v) == dict:
                for key, val in v.items():
                    if key.split('/')[3] == 'safe' or key.split('/')[3] == 'finalized':
                        pass
                    elif val != test_card[i]['result']['batch'][k][key]:
                        errors.append(f"problems on inner dict i {i}, k {k}, key {key}")
                        print(f"problems on inner dict i {i}, k {k}, key {key}")
            elif v != test_card[i]['result']['batch'][k]:
                errors.append(f"problems on outer dict i {i}, k {k}")
                print(f"problems on outter dict i {i}, k {k}")


    test_plugeth = test_plugeth[1:]
    control_plugeth = control_plugeth[1:]

    for i, item in enumerate(control_plugeth):
        for k,v in item['params']['result'].items():
            if type(v) == dict:
                for key, val in v.items():
                    if val != test_plugeth[i]['params']['result'][k][key]:
                        errors.append(f"problems on inner dict i {i}, k {k}, key {key}")
                        print(f"problems on inner dict i {i}, k {k}, key {key}")

            if v!= test_plugeth[i]['params']['result'][k]:
                errors.append(f"problems on outer dict i {i}, k {k}")
                print(f"problems on outer dict i {i}, k {k}")
    if errors:
        raise AssertionError("\n".join(errors))

    print("no problems found")

def main():
    test_cardinal()

if __name__ == "__main__":
    main()