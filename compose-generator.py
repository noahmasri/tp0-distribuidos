import os
import yaml, sys

# receives the number of clients and the output file.
# creates a new file (or replaces an existing file) with
# <output_file> as name. This yaml file will create a 
# compose file with <client_number> clientes.
def generate_compose(output_file: str, client_number: int):
    compose = {
        'name': 'tp0',
        'services': {
            'server': {
                'container_name': 'server',
                'image': 'server:latest',
                'entrypoint': 'python3 /main.py',
                'environment': [
                    'PYTHONUNBUFFERED=1',
                    'LOGGING_LEVEL=DEBUG'
                ],
                'networks': ['testing_net']
            },
            'client1': {
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
        },
        'networks': {
            'testing_net': {
                'ipam': {
                    'driver': 'default',
                    'config': [
                        {'subnet': '172.25.125.0/24'}
                    ]
                }
            }
        }
    }

    with open(output_file, 'w') as file:
        yaml.dump(compose, file,sort_keys=False, default_flow_style=False, indent=2)
    # to-do:
    #  * use the client_num to create more than one client
    print(output_file)
    print(client_number)

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