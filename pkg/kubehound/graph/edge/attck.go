package edge

// AttckTacticID is the interface for the ATT&CK tactic ID.
type AttckTacticID string

var (
	// AttckTacticUndefined is the undefined ATT&CK tactic.
	AttckTacticUndefined AttckTacticID
	// AttckTacticInitialAccess is the ATT&CK tactic for initial access (TA0001).
	AttckTacticInitialAccess AttckTacticID = "TA0001"
	// AttckTacticExecution is the ATT&CK tactic for execution (TA0002).
	AttckTacticExecution AttckTacticID = "TA0002"
	// AttckTacticPersistence is the ATT&CK tactic for persistence (TA0003).
	AttckTacticPersistence AttckTacticID = "TA0003"
	// AttckTacticPrivilegeEscalation is the ATT&CK tactic for privilege escalation (TA0004).
	AttckTacticPrivilegeEscalation AttckTacticID = "TA0004"
	// AttckTacticCredentialAccess is the ATT&CK tactic for credential access (TA0006).
	AttckTacticCredentialAccess AttckTacticID = "TA0006"
	// AttckTacticDiscovery is the ATT&CK tactic for discovery (TA0007).
	AttckTacticDiscovery AttckTacticID = "TA0007"
	// AttckTacticLateralMovement is the ATT&CK tactic for lateral movement (TA0008).
	AttckTacticLateralMovement AttckTacticID = "TA0008"
)

// AttckTechniqueID is the interface for the ATT&CK technique ID.
type AttckTechniqueID string

var (
	// AttckTechniqueUndefined is the undefined ATT&CK technique.
	AttckTechniqueUndefined AttckTechniqueID
	// AttckTechniquePermissionGroupsDiscovery is the ATT&CK technique for permission groups discovery (T1069).
	AttckTechniquePermissionGroupsDiscovery AttckTechniqueID = "T1069"
	// AttckTechniqueValidAccounts is the ATT&CK technique for valid accounts (T1078).
	AttckTechniqueValidAccounts AttckTechniqueID = "T1078"
	// AttckTechniqueExploitationOfRemoteServices is the ATT&CK technique for exploitation of remote services (T1210).
	AttckTechniqueExploitationOfRemoteServices AttckTechniqueID = "T1210"
	// AttckTechniqueStealApplicationAccessTokens is the ATT&CK technique for stealing application access tokens (T1528).
	AttckTechniqueStealApplicationAccessTokens AttckTechniqueID = "T1528"
	// AttckTechniqueUnsecuredCredentials is the ATT&CK technique for unsecured credentials (T1552).
	AttckTechniqueUnsecuredCredentials AttckTechniqueID = "T1552"
	// AttckTechniqueCreateOrModifySystemProcessContainerService is the ATT&CK technique for creating or modifying a system process container service (T1543.005).
	AttckTechniqueCreateOrModifySystemProcessContainerService AttckTechniqueID = "T1543.005"
	// AttckTechniqueContainerAdministrationCommand is the ATT&CK technique for container administration command (T1609).
	AttckTechniqueContainerAdministrationCommand AttckTechniqueID = "T1609"
	// AttckTechniqueDeployContainer is the ATT&CK technique for deploying a container (T1610).
	AttckTechniqueDeployContainer AttckTechniqueID = "T1610"
	// AttckTechniqueEscapeToHost is the ATT&CK technique for escaping to the host (T1611).
	AttckTechniqueEscapeToHost AttckTechniqueID = "T1611"
	// AttckTechniqueContainerAndResourceDiscovery is the ATT&CK technique for container and resource discovery (T1613).
	AttckTechniqueContainerAndResourceDiscovery AttckTechniqueID = "T1613"
)
