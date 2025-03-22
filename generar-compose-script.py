import sys
import yaml

def getComposeContentClient(client_id):
  return {
    'container_name': f'client{client_id}',
    'image': 'client:latest',
    'entrypoint': '/client',
    'environment': [f'CLI_ID={client_id}'],
    'networks': ['testing_net'],
    'depends_on': ['server'],
    'volumes': [''
      './client/config.yaml:/config.yaml',
      './.data:/.data',
    ]
  }

def getComposeContent(clients_amount):
  config = {
    'name': 'tp0',
    'services': {
      'server': {
        'container_name': 'server',
        'image': 'server:latest',
        'entrypoint': 'python3 /main.py',
        'environment': ['PYTHONUNBUFFERED=1'],
        'networks': ['testing_net'],
        'volumes': ['./server/config.ini:/config.ini']
      }
    },
    'networks': {
      'testing_net': {
        'ipam': {
          'driver': 'default',
          'config': [{'subnet': '172.25.125.0/24'}]
        }
      }
    } 
  }

  for i in range(1, clients_amount + 1):
      new_client = getComposeContentClient(i)
      config['services'][f'client{i}'] = new_client

  return config

def generate_docker_compose(output_file, clients_amount):
  compose_content = getComposeContent(clients_amount)
  
  with open(output_file, 'w') as f:
     yaml.dump(compose_content, f, default_flow_style=False, sort_keys=False)

def main():
  if len(sys.argv) != 3 or not sys.argv[2].isnumeric():
    print("Incorrect parameters. The correct use is: ./generar-compose.sh <output_file> <clients_amount>")
    sys.exit(1)

  output_file = sys.argv[1]
  clients_amount = int(sys.argv[2])

  generate_docker_compose(output_file, clients_amount)

main()