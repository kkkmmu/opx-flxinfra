COMPS=portd

IPCS=portd

all: ipc exe 

exe: $(COMPS)
	 $(foreach f,$^, make -C $(f) exe;)

ipc: $(IPCS)
	 $(foreach f,$^, make -C $(f) ipc;)

clean: $(COMPS)
         $(foreach f,$^, make -C $(f) clean;)
