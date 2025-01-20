[jenkinsmaster]
{{ .Host }} ansible_user={{ .User }} ansible_ssh_private_key_file={{ .PrivateKey }} ansible_port={{ .Port }}
