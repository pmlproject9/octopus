package constants

const (
	LabelSourceName      = "lighthouse.submariner.io/sourceName"
	LabelSourceNamespace = "lighthouse.submariner.io/sourceNamespace"
	LabelSourceClusterID = "lighthouse.submariner.io/sourceCluster"
	SubmarinerOperator   = "submariner-operator"
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
