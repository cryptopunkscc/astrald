# lab.story — the astrald lab, built in one netsim simulation.
# Result: a single stage with two nodes running astrald and a Qwen Code
# operator installed on node1.
add-vm --hostname node1
add-vm --hostname node2
install-astrald
install-qwen-code --vm node1 --create-user
