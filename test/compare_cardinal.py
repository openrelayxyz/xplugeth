import json

def test_cardinal():
    with open('resources/control_card_data.json', 'r') as cf:
        control = json.load(cf)

    with open('test_card_data.json', 'r') as tf:
        test = json.load(tf)

    for i, item in enumerate(control):
        for k, v in item['result']['batch'].items():
            if type(v) == dict:
                for key, val in v.items():
                    if key.split('/')[3] == 'safe' or key.split('/')[3] == 'finalized':
                        pass
                    elif val != test[i]['result']['batch'][k][key]:
                        print(f"problems on inner dict i {i}, k {k}, key {key}")
            elif v != test[i]['result']['batch'][k]:
                print(f"problems on outter dict i {i}, k {k}")
    print("no problems found")


def main():
    test_cardinal()

if __name__ == "__main__":
    main()