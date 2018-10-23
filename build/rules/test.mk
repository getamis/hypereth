coverage.txt:
	@touch $@

test: coverage.txt FORCE
	@for d in `go list ./... | grep -v vendor`; do		\
		go test -v -coverprofile=profile.out -covermode=atomic $$d;	\
		if [ $$? -eq 0 ]; then						\
			echo "\033[32mPASS\033[0m:\t$$d";			\
			if [ -f profile.out ]; then				\
				cat profile.out >> coverage.txt;		\
				rm profile.out;					\
			fi							\
		else								\
			echo "\033[31mFAIL\033[0m:\t$$d";			\
			exit -1;						\
		fi								\
	done;
