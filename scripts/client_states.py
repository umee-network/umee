#!/usr/bin/env python

from http import client
import sys
import json
import subprocess
from dateutil.parser import parse


page = 1
clients = []

while True:
  output = subprocess.check_output(['umeed', 'q', 'ibc', 'client', 'states', '--output', 'json', '--node', sys.argv[1], '--page', str(page)])
  data = json.loads(output.decode('utf-8'))

  clients.extend(data['client_states'])

  if not data['pagination']['next_key'] or not data['pagination']['total']: 
    break
  if data['pagination']['total'] == '0':
    break

  page += 1

# get the block time
output = subprocess.check_output(["umeed", "q", "ibc", "connection", "connections", "--output", "json", "--node", sys.argv[1]])

connection = json.loads(output.decode('utf-8'))

current_block_number = connection['height']['revision_height']
output = subprocess.check_output(["umeed", "q", "block", current_block_number, "--node", sys.argv[1]])
current_block = json.loads(output.decode('utf-8'))
current_block_time = parse(current_block['block']['header']['time'])

print(f"Current block time {str(current_block_time)} and number {current_block_number}")
print("")

for client_data in clients:
  print(client_data['client_id'])
  client_id = client_data['client_id']
  revision_height = client_data['client_state']['latest_height']['revision_height']
  revision_number = client_data['client_state']['latest_height']['revision_number']
  trusting_period = client_data['client_state']['trusting_period']

  # now get the time from the block at the revision height
  # and compare to the time at the current block
  print("client_id: " + client_id)
  print("chain_id: " + client_data['client_state']['chain_id'])
  print("revison: " + revision_number)
  print("revison height: " + revision_height)
  print('trusting period: ' + trusting_period)

  output = subprocess.check_output(['umeed', 'q', 'ibc', 'client', 'consensus-state', client_id,  f'{revision_number}-{revision_height}', '--node', sys.argv[1], '--output', 'json'])
  block = json.loads(output.decode('utf-8'))
  block_time = parse(block['consensus_state']['timestamp'])

  print("RPC endpoint block time: " + str(current_block_time))
  print("consensus block time: " + str(block_time))

  difference = (current_block_time - block_time).total_seconds()
  trust_period_int = int("".join(filter(str.isdigit, trusting_period)))

  if difference > int(trust_period_int):
    print(f"ERROR: Trusting period {trust_period_int} exceeded at {difference} seconds")
  else:
    print(f"{client_id} within trusting period {trust_period_int} with {difference}")

  print("")



