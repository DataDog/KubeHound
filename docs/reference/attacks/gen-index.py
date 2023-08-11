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
      docsConfig = yaml.safe_load(contents[startIndex+len(COMMENT_PREFIX):contents.find(COMMENT_SUFFIX)])
      attackTacticId, attackTacticName = docsConfig["mitreAttackTactic"].split(' - ')
      attackTechniqueId, attackTechniqueName = docsConfig["mitreAttackTechnique"].split(' - ')
      table += f'| [{docsConfig["id"]}](./{file}) | {docsConfig["name"]} | {attackTechniqueName} | { attackTacticName} | \n'
    else:
      print(f"WARNING: {file} does not have a docs config")
      
      
with open('index.md', 'w') as f:
  f.write(table)
      