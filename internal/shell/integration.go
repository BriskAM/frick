package shell

import "fmt"

// GetIntegrationScript returns the shell integration script that registers the frick function.
func GetIntegrationScript() string {
	return `
# frick shell integration
if [ -n "$ZSH_VERSION" ]; then
	frick() {
		local last_exit=$?
		local last_cmd=$(fc -ln -1)
		# Trim leading/trailing whitespace
		last_cmd=$(echo "$last_cmd" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
		
		# Skip if last command is empty or is frick itself
		if [ -z "$last_cmd" ] || [ "$last_cmd" = "frick" ] || [[ "$last_cmd" =~ ^frick[[:space:]] ]]; then
			return
		fi
		
		local corrected_cmd
		corrected_cmd=$(command frick run --cmd "$last_cmd" --exit "$last_exit")
		
		if [ -n "$corrected_cmd" ]; then
			print -s "$corrected_cmd"
			eval "$corrected_cmd"
		fi
	}
elif [ -n "$BASH_VERSION" ]; then
	frick() {
		local last_exit=$?
		local last_cmd=$(fc -ln -1)
		# Trim leading/trailing whitespace
		last_cmd=$(echo "$last_cmd" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
		
		# Skip if last command is empty or is frick itself
		if [ -z "$last_cmd" ] || [ "$last_cmd" = "frick" ] || [[ "$last_cmd" =~ ^frick[[:space:]] ]]; then
			return
		fi
		
		local corrected_cmd
		corrected_cmd=$(command frick run --cmd "$last_cmd" --exit "$last_exit")
		
		if [ -n "$corrected_cmd" ]; then
			# Clean up history: delete frick and the failed command
			if [[ "$HISTCMD" =~ ^[0-9]+$ ]]; then
				history -d $(( HISTCMD - 1 )) 2>/dev/null
				history -d $(( HISTCMD )) 2>/dev/null
			fi
			history -s "$corrected_cmd"
			eval "$corrected_cmd"
		fi
	}
else
	# Fallback for other POSIX shells
	frick() {
		local last_exit=$?
		local last_cmd=$(fc -ln -1 2>/dev/null || history 1 2>/dev/null | sed -e "s/^[ ]*[0-9]*[ ]*//")
		last_cmd=$(echo "$last_cmd" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
		
		if [ -z "$last_cmd" ] || [ "$last_cmd" = "frick" ] || [[ "$last_cmd" =~ ^frick[[:space:]] ]]; then
			return
		fi
		
		local corrected_cmd
		corrected_cmd=$(command frick run --cmd "$last_cmd" --exit "$last_exit")
		
		if [ -n "$corrected_cmd" ]; then
			eval "$corrected_cmd"
		fi
	}
fi
`
}

// PrintIntegrationScript outputs the script to stdout.
func PrintIntegrationScript() {
	fmt.Print(GetIntegrationScript())
}
