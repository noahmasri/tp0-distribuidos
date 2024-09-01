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
            'LOGGING_LEVEL=INFO'
        ],
        'networks': ['testing_net'],
        'volumes': ['./server/config.ini:/config.ini']
    }
    return compose

def add_client(compose: dict[str, Any], client_id: int):
    compose['services'][f'client{client_id}']= {
        'container_name': f'client{client_id}',
        'image': 'client:latest',
        'entrypoint': '/client',
        'environment': [
            f'CLI_ID={client_id}',
            'CLI_LOG_LEVEL=INFO'
        ],
        'networks': ['testing_net'],
        'depends_on': ['server'],
        'volumes': ['./client/config.yaml:/config.yaml']
    }
    return compose

def add_all_clients(compose: dict[str, Any], client_number: int):
    for i in range(1, client_number + 1):
        compose = add_client(compose, i)

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
    compose = add_all_clients(compose, client_number)
    compose = add_testing_network(compose)

    with open(output_file, 'w') as file:
        yaml.dump(compose, file,sort_keys=False, default_flow_style=False)

def main():
    if len(sys.argv) != 3:
        print("run with: python3 compose-generator.py <output_file> <client_number>")
        return
    
    try:
        client_number = int(sys.argv[2])
        generate_compose(sys.argv[1], client_number)
    except ValueError:
        print(f"Error: second argument has to be a number (amount of clients desired)")

if __name__ == '__main__':
    main()