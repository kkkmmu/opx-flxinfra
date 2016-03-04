namespace go sysd
typedef i32 int
typedef i16 uint16
struct ComponentLoggingConfig {
	1 : string Module
	2 : string Level
}
struct SystemLoggingConfig {
	1 : string SRLogger
	2 : string SystemLogging
}
service SYSDServices {
	bool CreateComponentLoggingConfig(1: ComponentLoggingConfig config);
	bool UpdateComponentLoggingConfig(1: ComponentLoggingConfig origconfig, 2: ComponentLoggingConfig newconfig, 3: list<bool> attrset);
	bool DeleteComponentLoggingConfig(1: ComponentLoggingConfig config);

	bool CreateSystemLoggingConfig(1: SystemLoggingConfig config);
	bool UpdateSystemLoggingConfig(1: SystemLoggingConfig origconfig, 2: SystemLoggingConfig newconfig, 3: list<bool> attrset);
	bool DeleteSystemLoggingConfig(1: SystemLoggingConfig config);

}