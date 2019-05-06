import logging
import json
import os
from jsonschema import validate

CONFIG_FILE = '/etc/nno/config.json'
CONFIG_SCHEMA = {
    # "$schema": "http://json-schema.org/schema#",
    # "$id": "http://mellanoxoperator.com/schemas/configuration.json",
    "type": "array", "items": {
        "type": "object",
        "properties": {
            "pfName": {"type": "string", "minLength": 3},
            "numOfVfs": {"type": "integer", "minimum": 0},
            "totalVfs": {"type": "integer", "minimum": 0},
            "linkType": {"type": "string", "enum": ["ETH", "IB"]},
        },
        "required": ["pfName"],
        "additionalProperties": False
    }
}


def get_pci_addrs_from_net_dev_name(net_dev):
    return os.readlink("/sys/class/net/%s/device" % net_dev)[9:]


def apply_configs(conf):
    pci_addrs = get_pci_addrs_from_net_dev_name(conf["pfName"])
    numOfVfs = conf.get("numOfVfs")
    if numOfVfs is not None:
        os.system("echo %s > /sys/bus/pci/devices/%s/sriov_numvfs" % (numOfVfs, pci_addrs))

    totalVfs = conf.get("totalVfs")
    if totalVfs is not None:
        os.system("mstconfig -d %s -y  set NUM_OF_VFS=%s" % (pci_addrs, totalVfs))

    linkType = conf.get("linkType")
    if linkType is not None:
        os.system("mstconfig -d %s -y  set LINK_TYPE_P1=%s" % (pci_addrs, linkType))


def main():
    logging.basicConfig(level=logging.DEBUG,
                        format='%(asctime)-15s %(filename)s %(lineno)d [%(levelname)s]: %(message)s')
    logger = logging.getLogger(__file__)
    configs = {}
    logger.debug("reading configuration file %s" % CONFIG_FILE)
    try:
        with open(CONFIG_FILE) as configsFile:
            # Read the configurations and validate them
            configs = json.load(configsFile)
            validate(configs, CONFIG_SCHEMA)

            # run commands
            for conf in configs:
                apply_configs(conf)
    except Exception as err:
        logger.error(err)
        exit(1)


if __name__ == '__main__':
    main()

