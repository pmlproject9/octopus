package constants

const (
	LabelSourceName      = "octopus.io/sourceName"
	LabelSourceNamespace = "octopus.io/sourceNamespace"
	LabelSourceClusterID = "octopus.io/sourceClusterID"
)

const (
	IPSetAllServicesCidr = "OCTOPUS-MNP-ALL-SERVICECIDR"
	IPSetPodPrefix       = "OCTOPUS-POD-"
	IPSetServicePrefix   = "OCTOPUS-SVC-"
)

const (
	IPTablesFilter                 = "filter"
	IPTablesChainOCTOPUSFORWARD    = "OCTOPUS-FORWARD"
	IPTablesChainOCTOPUSMNPFORWARD = "OCTOPUS-MNP-FORWARD"
	IPTablesChainOCTOPUSFWPREFIX   = "OCTOPUS-FW-"
)
