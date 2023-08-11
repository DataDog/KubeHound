import glob
import yaml

COMMENT_PREFIX = '<!--'
COMMENT_SUFFIX = '-->'


table = """---
hide:
  - toc
---

# Attack Reference

|   ID   | Name | MITRE ATT&CK Technique | MITRE ATT&CK Tactic |
| :----: | :--: | :-----------------: | :--------------------: |
"""

for file in sorted(glob.glob('*.md')):
  if file == 'index.md':
    continue
  
  with open(file, 'r') as f:
    contents = f.read()
    startIndex = contents.find(COMMENT_PREFIX)
    if startIndex >= 0:
      print("Parsing", file)
      docs_config = yaml.safe_load(contents[startIndex+len(COMMENT_PREFIX):contents.find(COMMENT_SUFFIX)])
      attackTacticId, attackTacticName = docs_config["mitreAttackTactic"].split(' - ')
      attackTechniqueId, attackTechniqueName = docs_config["mitreAttackTechnique"].split(' - ')
      table += f'| [{docs_config["id"]}](./{file}) | {docs_config["name"]} | {attackTechniqueName} | { attackTacticName} | \n'
    else:
      print(f"WARNING: {file} does not have a docs config")
      
      
with open('index.md', 'w') as f:
  f.write(table)
      