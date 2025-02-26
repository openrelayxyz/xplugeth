import json, logging
import pytest_check as check

logging.basicConfig(level=logging.INFO, format="%(levelname)s - %(message)s")

def check_blockupdates_values():
    try:
        with open('./resources/control_plugeth_data.json', 'r') as cf:
            control_plugeth = json.load(cf)

        with open('./resources/test_plugeth_data.json', 'r') as tf:
            test_plugeth = json.load(tf)
    except Exception as e:
        logging.error("error opening test and control files: {e}")
        raise

    # we have to trim the first item from each blockupdates list as they are subcription ids and have no payload
    test_plugeth = test_plugeth[1:]
    control_plugeth = control_plugeth[1:]

    for i, item in enumerate(control_plugeth):
        for k, v in item['params']['result'].items():
            if isinstance(v, dict):
                for key, val in v.items():
                    check.equal(
                        val, test_plugeth[i]['params']['result'][k][key],
                        f"Mismatch at inner dict blockupdates: index {i}, key '{k}', subkey '{key}'. "
                        f"Expected {val}, got {test_plugeth[i]['params']['result'][k][key]}"
                    )
            
            check.equal(
                    v, test_plugeth[i]['params']['result'][k],
                    f"Mismatch at outer dict blockupdates: index {i}, key '{k}'. "
                    f"Expected {v}, got {test_plugeth[i]['params']['result'][k]}"
            )


def check_cardinal_values():
    try:
        with open('./resources/control_card_data.json', 'r') as cf:
            control_card = json.load(cf)

        with open('./resources/test_card_data.json', 'r') as tf:
            test_card = json.load(tf)
    except Exception as e:
        logging.error("error opening test and control files: {e}")
        raise

    for i, item in enumerate(control_card):
        for k, v in item['result']['batch'].items():
            if isinstance(v,dict):
                for key, val in v.items():
                    check.equal(
                        val, test_card[i]['result']['batch'][k][key],
                        f"Mismatch at inner dict cardinal: index {i}, key '{k}', subkey '{key}'. "
                        f"Expected {val}, got {test_card[i]['result']['batch'][k][key]}"
                    )
            check.equal(
                    v, test_card[i]['result']['batch'][k],
                    f"Mismatch at outer dict cardinal: index {i}, key '{k}'. "
                    f"Expected {v}, got {test_card[i]['result']['batch'][k]}"
            )
