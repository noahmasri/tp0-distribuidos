import os
from typing import Any
import yaml, sys

def add_server(compose: dict[str, Any]):
    compose['services']['server']= {
        'container_name': 'server',
        'image': 'server:latest',
        'entrypoint': 'python3 /main.py',
        'environment': [
            'PYTHONUNBUFFERED=1',
            'LOGGING_LEVEL=DEBUG'
        ],
        'networks': ['testing_net']
    }
    return compose

def add_client(compose: dict[str, Any]):
    compose['services']['client1']= {
        'container_name': 'client1',
        'image': 'client:latest',
        'entrypoint': '/client',
        'environment': [
            'CLI_ID=1',
            'CLI_LOG_LEVEL=DEBUG'
        ],
        'networks': ['testing_net'],
        'depends_on': ['server']
    }
    return compose

def add_testing_network(compose: dict[str, Any]):
    compose['networks']['testing_net'] = {
        'ipam': {
            'driver': 'default',
            'config': [
                {'subnet': '172.25.125.0/24'}
            ]
        }
    }
    return compose
"""
receives the number of clients and the output file.
creates a new file (or replaces an existing file) with
<output_file> as name. This yaml file will create a 
compose file with <client_number> clientes.
"""
def generate_compose(output_file: str, client_number: int):
    compose = {
        'name': 'tp0',
        'services': {
        },
        'networks': {
        }
    }

    compose = add_server(compose)
    compose = add_client(compose)
    compose = add_testing_network(compose)

    with open(output_file, 'w') as file:
        yaml.dump(compose, file,sort_keys=False, default_flow_style=False)
    # to-do:
    #  * use the client_num to create more than one client

def main():
    if len(sys.argv) != 3:
        print("run with: python3 compose-generator.py <output_file> <client_number>")
        return
    
    # see if i keep it or not
    _, extension = os.path.splitext(sys.argv[1])
    print(extension)
    if extension != ".yaml":
        print("output file has to be a yaml file")
        return

    generate_compose(sys.argv[1], sys.argv[2])


if __name__ == '__main__':
    main()