import glob
import yaml

COMMENT_PREFIX = '<!--'
COMMENT_SUFFIX = '-->'


table = """---
hide:
  - toc
---

# Attack Reference

All edges in the KubeHound graph represent attacks with a net "improvement" in an attacker's position or a lateral movement opportunity.

!!! note

    For instance, an assume role or ([IDENTITY_ASSUME](./IDENTITY_ASSUME.md)) is considered as an attack.

|   ID   | Name | MITRE ATT&CK Technique | MITRE ATT&CK Tactic | Coverage |
| :----: | :--: | :-----------------: | :--------------------: | :------: |
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
      # Extract and format MITRE ATT&CK Tactic
      attackTacticName = "N/A"
      if docsConfig["mitreAttackTactic"] != "N/A":
        attackTacticId, attackTacticName = docsConfig["mitreAttackTactic"].split(' - ')
      # Extract and format MITRE ATT&CK Technique
      attackTechniqueName = "N/A"
      if docsConfig["mitreAttackTechnique"] != "N/A":
        attackTechniqueId, attackTechniqueName = docsConfig["mitreAttackTechnique"].split(' - ')
      # Extract coverage
      coverage = "Full"
      if "coverage" in docsConfig:
        coverage = docsConfig["coverage"]
      # Generate table row
      table += f'| [{docsConfig["id"]}](./{file}) | {docsConfig["name"]} | {attackTechniqueName} | { attackTacticName} | {coverage} |\n'
    else:
      print(f"WARNING: {file} does not have a docs config")
      
      
with open('index.md', 'w') as f:
  f.write(table)
      