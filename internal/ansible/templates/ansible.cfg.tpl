[defaults]
inventory = {{ .InventoryFile }}
host_key_checking = False
forks = {{ .Forks }}
pipelining = True

[ssh_connection]
pipelining = True
