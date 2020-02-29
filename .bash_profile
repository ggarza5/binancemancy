# .bash_profile

# Get the aliases and functions
if [ -f ~/.bashrc ]; then
	. ~/.bashrc
fi

# User specific environment and startup programs

PATH=$PATH:$HOME/.local/bin:$HOME/bin
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:/home/ec2-user/go/bin
export PATH

# export HOME=/home/ec2-user/go
