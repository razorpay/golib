echo 'exec echo ${GIT_TOKEN}' > /tmp/askpass.sh && chmod +x /tmp/askpass.sh
export GIT_ASKPASS=/tmp/askpass.sh
export CGO_ENABLED=1
# Function to check if a directory contains a go.mod file
has_go_mod() {
  [ -e "go.mod" ]
}

# get the go version from the go version command 1.18/1.19/1.20/1.21
go_version=`go version | cut -c 14-17`
cat /etc/passwd
# Recursive function to run tests in all directories
run_tests_recursive() {
  for d in */; do
    cd "$d"
    if has_go_mod $d; then
      mod_version=`sed -En 's/^go[[:space:]]+([[:digit:].]+)$/\1/p' go.mod`;
      printf '%s\n' "$d module go version - $go_version"
      if [ "$go_version" = "$mod_version" ]; then
        printf '%s\n' "----------- Testing module - $d -------------------"
        make run-test
        if [ $? -eq 0 ]; then
            echo "$d tests ran successfully"
        else
            exit 1
        fi
      else
        printf '%s\n' "------- skipping module - $d ---------------"
      fi
    else
      run_tests_recursive "$d"
    fi
    cd ..
  done
}

run_tests_recursive
