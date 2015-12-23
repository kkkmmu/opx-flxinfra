RM=rm -f
RMFORCE=rm -rf
DESTDIR=$(SR_CODE_BASE)/snaproute/src/out/bin
GENERATED_IPC=$(SR_CODE_BASE)/generated/src
PARAMSDIR=$(DESTDIR)/params
MKDIR=mkdir -p
IPC_GEN_CMD=thrift 
SRCS=porthandler.go portmain.go
IPC_SRCS=portd.thrift
IPC_SVC_NAME=portdServices
COMP_NAME=portd
all:ipc exe install
ipc:
	 $(IPC_GEN_CMD) --gen go -out $(GENERATED_IPC) $(IPC_SRCS)

exe: $(SRCS)
	 go build -o $(DESTDIR)/$(COMP_NAME) -ldflags="$(GOLDFLAGS)" $(SRCS)

install:
	 @$(MKDIR) $(PARAMSDIR)
	 @install params/portd.json $(PARAMSDIR)/

guard:
ifndef SR_CODE_BASE
	 $(error SR_CODE_BASE is not set)
endif

clean:guard
	 $(RM) $(DESTDIR)/$(COMP_NAME) 
	 $(RMFORCE) $(GENERATED_IPC)/$(COMP_NAME)
