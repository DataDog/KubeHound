from urllib.request import urlopen, Request
import json
import os
import sys

with open(sys.argv[1], 'r') as f:
  licences = f.read().splitlines()
  
github_users_to_names = {}
GITHUB_API_TOKEN = os.getenv('GITHUB_AUTH_TOKEN')
if not GITHUB_API_TOKEN:
  raise ValueError("Please set GITHUB_AUTH_TOKEN, you can generate it from https://github.com/settings/tokens")

def get_github_user_full_name(username):
  request = Request(f'https://api.github.com/users/{username}', headers={
    'Authorization': f'Bearer {GITHUB_API_TOKEN}'
  })
  with urlopen(request) as response:
    if response.status != 200:
      raise RuntimeError(f'Failed to get user info for user {username} from the GitHub API')
    parsed_response = json.loads(response.read())
    name = parsed_response.get('name')
    if name is None:
      name = username # fallback to the username
      
    return name
    
for row in licences:
  module, url, license = row.split(',')
  if url.startswith("https://github.com/"):
    username = url.split('/')[3]
    if username not in github_users_to_names:
      github_users_to_names[username] = get_github_user_full_name(username)
    name = github_users_to_names[username]
  else:
    name = username
    
  print(f'{module},{url},{license},{github_users_to_names[username]}')
  sys.stdout.flush()

      
      
