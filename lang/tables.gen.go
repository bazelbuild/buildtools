// Generated file, do not edit.
package lang

import buildpb "github.com/bazelbuild/buildtools/build_proto"

var TypeOf = map[string]buildpb.Attribute_Discriminator{
	"$aar_embedded_jars_extractor":	buildpb.Attribute_LABEL,
	"$aar_native_libs_zip_creator":	buildpb.Attribute_LABEL,
	"$aar_resources_extractor":	buildpb.Attribute_LABEL,
	"$adb":	buildpb.Attribute_LABEL,
	"$adb_static":	buildpb.Attribute_LABEL,
	"$android_resources_busybox":	buildpb.Attribute_LABEL,
	"$android_runtest":	buildpb.Attribute_LABEL,
	"$build_incremental_dexmanifest":	buildpb.Attribute_LABEL,
	"$build_java8_legacy_dex":	buildpb.Attribute_LABEL,
	"$build_split_manifest":	buildpb.Attribute_LABEL,
	"$cc_toolchain_split":	buildpb.Attribute_LABEL,
	"$cc_toolchain_type":	buildpb.Attribute_STRING,
	"$child_configuration_dummy":	buildpb.Attribute_LABEL,
	"$collect_cc_coverage":	buildpb.Attribute_LABEL,
	"$collect_coverage_script":	buildpb.Attribute_LABEL,
	"$config_dependencies":	buildpb.Attribute_LABEL_LIST,
	"$databinding_annotation_processor":	buildpb.Attribute_LABEL,
	"$databinding_exec":	buildpb.Attribute_LABEL,
	"$def_parser":	buildpb.Attribute_LABEL,
	"$desugar":	buildpb.Attribute_LABEL,
	"$desugar_java8_extra_bootclasspath":	buildpb.Attribute_LABEL,
	"$desugared_java8_legacy_apis":	buildpb.Attribute_LABEL,
	"$dex_list_obfuscator":	buildpb.Attribute_LABEL,
	"$dexbuilder":	buildpb.Attribute_LABEL,
	"$dexbuilder_after_proguard":	buildpb.Attribute_LABEL,
	"$dexmerger":	buildpb.Attribute_LABEL,
	"$dexsharder":	buildpb.Attribute_LABEL,
	"$dummy_lib":	buildpb.Attribute_LABEL,
	"$empty_snapshot_fs":	buildpb.Attribute_LABEL,
	"$emulator_arm":	buildpb.Attribute_LABEL,
	"$emulator_x86":	buildpb.Attribute_LABEL,
	"$emulator_x86_bios":	buildpb.Attribute_LABEL,
	"$genrule_setup":	buildpb.Attribute_LABEL,
	"$googlemac_proto_compiler":	buildpb.Attribute_LABEL,
	"$googlemac_proto_compiler_support":	buildpb.Attribute_LABEL,
	"$grep_includes":	buildpb.Attribute_LABEL,
	"$host_jdk":	buildpb.Attribute_LABEL,
	"$idlclass":	buildpb.Attribute_LABEL,
	"$implicit_tests":	buildpb.Attribute_LABEL_LIST,
	"$import_deps_checker":	buildpb.Attribute_LABEL,
	"$incremental_install":	buildpb.Attribute_LABEL,
	"$incremental_split_stub_application":	buildpb.Attribute_LABEL,
	"$incremental_stub_application":	buildpb.Attribute_LABEL,
	"$instrumentation_test_check":	buildpb.Attribute_LABEL,
	"$interface_library_builder":	buildpb.Attribute_LABEL,
	"$is_executable":	buildpb.Attribute_BOOLEAN,
	"$j2objc_dead_code_pruner":	buildpb.Attribute_LABEL,
	"$jacocorunner":	buildpb.Attribute_LABEL,
	"$java8_legacy_dex":	buildpb.Attribute_LABEL,
	"$java_toolchain":	buildpb.Attribute_LABEL,
	"$jvm":	buildpb.Attribute_LABEL,
	"$launcher":	buildpb.Attribute_LABEL,
	"$libtool":	buildpb.Attribute_LABEL,
	"$link_dynamic_library_tool":	buildpb.Attribute_LABEL,
	"$merge_dexzips":	buildpb.Attribute_LABEL,
	"$mksd":	buildpb.Attribute_LABEL,
	"$proguard_whitelister":	buildpb.Attribute_LABEL,
	"$protobuf_well_known_types":	buildpb.Attribute_LABEL_LIST,
	"$py_toolchain_type":	buildpb.Attribute_STRING,
	"$python2to3":	buildpb.Attribute_LABEL,
	"$resource_extractor":	buildpb.Attribute_LABEL,
	"$robolectric_implicit_classpath":	buildpb.Attribute_LABEL_LIST,
	"$sdk_path":	buildpb.Attribute_LABEL,
	"$shuffle_jars":	buildpb.Attribute_LABEL,
	"$strip_resources":	buildpb.Attribute_LABEL,
	"$stubify_manifest":	buildpb.Attribute_LABEL,
	"$test_entry_point":	buildpb.Attribute_LABEL,
	"$test_runtime":	buildpb.Attribute_LABEL_LIST,
	"$test_setup_script":	buildpb.Attribute_LABEL,
	"$test_wrapper":	buildpb.Attribute_LABEL,
	"$testing_shbase":	buildpb.Attribute_LABEL,
	"$testsupport":	buildpb.Attribute_LABEL,
	"$tools_repository":	buildpb.Attribute_STRING,
	"$unified_launcher":	buildpb.Attribute_LABEL,
	"$whitelist_allow_deps_without_srcs":	buildpb.Attribute_LABEL,
	"$whitelist_android_device":	buildpb.Attribute_LABEL,
	"$whitelist_config_feature_flag":	buildpb.Attribute_LABEL,
	"$whitelist_disabling_parse_headers_and_layering_check_allowed":	buildpb.Attribute_LABEL,
	"$whitelist_export_deps":	buildpb.Attribute_LABEL,
	"$whitelist_loose_header_check_allowed_in_toolchain":	buildpb.Attribute_LABEL,
	"$xcrunwrapper":	buildpb.Attribute_LABEL,
	"$xml_generator_script":	buildpb.Attribute_LABEL,
	"$xml_writer":	buildpb.Attribute_LABEL,
	"$xvfb_support":	buildpb.Attribute_LABEL,
	"$zip_filter":	buildpb.Attribute_LABEL,
	"$zipper":	buildpb.Attribute_LABEL,
	":action_listener":	buildpb.Attribute_LABEL_LIST,
	":alias":	buildpb.Attribute_LABEL,
	":android_sdk":	buildpb.Attribute_LABEL,
	":aspect_proto_toolchain_for_javalite":	buildpb.Attribute_LABEL,
	":bytecode_optimizer":	buildpb.Attribute_LABEL,
	":cc_toolchain":	buildpb.Attribute_LABEL,
	":coverage_report_generator":	buildpb.Attribute_LABEL,
	":coverage_support":	buildpb.Attribute_LABEL,
	":csfdo_profile":	buildpb.Attribute_LABEL,
	":default_malloc":	buildpb.Attribute_LABEL,
	":extra_proguard_specs":	buildpb.Attribute_LABEL_LIST,
	":fdo_optimize":	buildpb.Attribute_LABEL,
	":fdo_prefetch_hints":	buildpb.Attribute_LABEL,
	":fdo_profile":	buildpb.Attribute_LABEL,
	":java_launcher":	buildpb.Attribute_LABEL,
	":java_plugins":	buildpb.Attribute_LABEL_LIST,
	":lcov_merger":	buildpb.Attribute_LABEL,
	":legacy_main_dex_list_generator":	buildpb.Attribute_LABEL,
	":libc_top":	buildpb.Attribute_LABEL,
	":proguard":	buildpb.Attribute_LABEL,
	":proto_compiler":	buildpb.Attribute_LABEL,
	":py_interpreter":	buildpb.Attribute_LABEL,
	":run_under":	buildpb.Attribute_LABEL,
	":target_libc_top":	buildpb.Attribute_LABEL,
	":xcode_config":	buildpb.Attribute_LABEL,
	":xfdo_profile":	buildpb.Attribute_LABEL,
	":zipper":	buildpb.Attribute_LABEL,
	"aapt":	buildpb.Attribute_LABEL,
	"aapt2":	buildpb.Attribute_LABEL,
	"aar":	buildpb.Attribute_LABEL,
	"absolute_path_profile":	buildpb.Attribute_STRING,
	"actual":	buildpb.Attribute_LABEL,
	"adb":	buildpb.Attribute_LABEL,
	"additional_linker_inputs":	buildpb.Attribute_LABEL_LIST,
	"aidl":	buildpb.Attribute_LABEL,
	"aidl_lib":	buildpb.Attribute_LABEL,
	"aliases":	buildpb.Attribute_STRING_LIST,
	"all_files":	buildpb.Attribute_LABEL,
	"allowed_values":	buildpb.Attribute_STRING_LIST,
	"alwayslink":	buildpb.Attribute_BOOLEAN,
	"android_jar":	buildpb.Attribute_LABEL,
	"annotations_jar":	buildpb.Attribute_LABEL,
	"api_level":	buildpb.Attribute_INTEGER,
	"api_version":	buildpb.Attribute_INTEGER,
	"apkbuilder":	buildpb.Attribute_LABEL,
	"apksigner":	buildpb.Attribute_LABEL,
	"applicable_licenses":	buildpb.Attribute_LABEL_LIST,
	"application_resources":	buildpb.Attribute_LABEL,
	"ar_files":	buildpb.Attribute_LABEL,
	"archives":	buildpb.Attribute_LABEL_LIST,
	"args":	buildpb.Attribute_STRING_LIST,
	"as_files":	buildpb.Attribute_LABEL,
	"assets":	buildpb.Attribute_LABEL_LIST,
	"assets_dir":	buildpb.Attribute_STRING,
	"avoid_deps":	buildpb.Attribute_LABEL_LIST,
	"binary_type":	buildpb.Attribute_STRING,
	"blacklisted_protos":	buildpb.Attribute_LABEL_LIST,
	"bootclasspath":	buildpb.Attribute_LABEL_LIST,
	"bootstrap_template":	buildpb.Attribute_LABEL,
	"build_file":	buildpb.Attribute_STRING,
	"build_file_content":	buildpb.Attribute_STRING,
	"build_setting_default":	buildpb.Attribute_LABEL,
	"build_tools_version":	buildpb.Attribute_STRING,
	"buildpar":	buildpb.Attribute_BOOLEAN,
	"bundle_loader":	buildpb.Attribute_LABEL,
	"cache":	buildpb.Attribute_INTEGER,
	"classpath_resources":	buildpb.Attribute_LABEL_LIST,
	"cmd":	buildpb.Attribute_STRING,
	"cmd_bash":	buildpb.Attribute_STRING,
	"cmd_bat":	buildpb.Attribute_STRING,
	"cmd_ps":	buildpb.Attribute_STRING,
	"command_line":	buildpb.Attribute_STRING,
	"compatible_javacopts":	buildpb.Attribute_STRING_LIST_DICT,
	"compatible_with":	buildpb.Attribute_LABEL_LIST,
	"compiler":	buildpb.Attribute_STRING,
	"compiler_files":	buildpb.Attribute_LABEL,
	"compiler_files_without_includes":	buildpb.Attribute_LABEL,
	"constraint_setting":	buildpb.Attribute_LABEL,
	"constraint_values":	buildpb.Attribute_LABEL_LIST,
	"constraints":	buildpb.Attribute_STRING_LIST,
	"copts":	buildpb.Attribute_STRING_LIST,
	"coverage_files":	buildpb.Attribute_LABEL,
	"coverage_tool":	buildpb.Attribute_LABEL,
	"cpu":	buildpb.Attribute_STRING,
	"cpu_constraints":	buildpb.Attribute_LABEL_LIST,
	"create_executable":	buildpb.Attribute_BOOLEAN,
	"crunch_png":	buildpb.Attribute_BOOLEAN,
	"custom_package":	buildpb.Attribute_STRING,
	"daemon":	buildpb.Attribute_BOOLEAN,
	"data":	buildpb.Attribute_LABEL_LIST,
	"debug_key":	buildpb.Attribute_LABEL,
	"debug_signing_keys":	buildpb.Attribute_LABEL_LIST,
	"debug_signing_lineage_file":	buildpb.Attribute_LABEL,
	"default":	buildpb.Attribute_LABEL,
	"default_applicable_licenses":	buildpb.Attribute_LABEL_LIST,
	"default_constraint_value":	buildpb.Attribute_STRING,
	"default_copts":	buildpb.Attribute_STRING_LIST,
	"default_deprecation":	buildpb.Attribute_STRING,
	"default_hdrs_check":	buildpb.Attribute_STRING,
	"default_ios_sdk_version":	buildpb.Attribute_STRING,
	"default_macos_sdk_version":	buildpb.Attribute_STRING,
	"default_package_metadata":	buildpb.Attribute_LABEL_LIST,
	"default_properties":	buildpb.Attribute_LABEL,
	"default_testonly":	buildpb.Attribute_BOOLEAN,
	"default_tvos_sdk_version":	buildpb.Attribute_STRING,
	"default_value":	buildpb.Attribute_STRING,
	"default_visibility":	buildpb.Attribute_STRING_LIST,
	"default_watchos_sdk_version":	buildpb.Attribute_STRING,
	"define_values":	buildpb.Attribute_STRING_DICT,
	"defines":	buildpb.Attribute_STRING_LIST,
	"densities":	buildpb.Attribute_STRING_LIST,
	"deploy_env":	buildpb.Attribute_LABEL_LIST,
	"deploy_manifest_lines":	buildpb.Attribute_STRING_LIST,
	"deprecation":	buildpb.Attribute_STRING,
	"deps":	buildpb.Attribute_LABEL_LIST,
	"deps_mapping":	buildpb.Attribute_LABEL_DICT_UNARY,
	"dex_shards":	buildpb.Attribute_INTEGER,
	"dexopts":	buildpb.Attribute_STRING_LIST,
	"distribs":	buildpb.Attribute_DISTRIBUTION_SET,
	"dwp_files":	buildpb.Attribute_LABEL,
	"dx":	buildpb.Attribute_LABEL,
	"dylibs":	buildpb.Attribute_LABEL_LIST,
	"dynamic_deps":	buildpb.Attribute_LABEL_LIST,
	"dynamic_runtime_lib":	buildpb.Attribute_LABEL,
	"enable_data_binding":	buildpb.Attribute_BOOLEAN,
	"enable_modules":	buildpb.Attribute_BOOLEAN,
	"entry_classes":	buildpb.Attribute_STRING_LIST,
	"exec_compatible_with":	buildpb.Attribute_LABEL_LIST,
	"exec_properties":	buildpb.Attribute_STRING_DICT,
	"exec_tools":	buildpb.Attribute_LABEL_LIST,
	"executable":	buildpb.Attribute_BOOLEAN,
	"exported_plugins":	buildpb.Attribute_LABEL_LIST,
	"exports":	buildpb.Attribute_LABEL_LIST,
	"exports_manifest":	buildpb.Attribute_TRISTATE,
	"expression":	buildpb.Attribute_STRING,
	"extclasspath":	buildpb.Attribute_LABEL_LIST,
	"extension_safe":	buildpb.Attribute_BOOLEAN,
	"extra_actions":	buildpb.Attribute_LABEL_LIST,
	"extra_srcs":	buildpb.Attribute_LABEL_LIST,
	"feature_flags":	buildpb.Attribute_LABEL_KEYED_STRING_DICT,
	"features":	buildpb.Attribute_STRING_LIST,
	"files":	buildpb.Attribute_LABEL_LIST,
	"fixtures":	buildpb.Attribute_LABEL_LIST,
	"flag_values":	buildpb.Attribute_LABEL_KEYED_STRING_DICT,
	"flaky":	buildpb.Attribute_BOOLEAN,
	"forcibly_disable_header_compilation":	buildpb.Attribute_BOOLEAN,
	"framework_aidl":	buildpb.Attribute_LABEL,
	"fulfills":	buildpb.Attribute_LABEL_LIST,
	"genclass":	buildpb.Attribute_LABEL_LIST,
	"generates_api":	buildpb.Attribute_BOOLEAN,
	"generator_function":	buildpb.Attribute_STRING,
	"generator_location":	buildpb.Attribute_STRING,
	"generator_name":	buildpb.Attribute_STRING,
	"has_services":	buildpb.Attribute_BOOLEAN,
	"hdrs":	buildpb.Attribute_LABEL_LIST,
	"header_compiler":	buildpb.Attribute_LABEL_LIST,
	"header_compiler_builtin_processors":	buildpb.Attribute_STRING_LIST,
	"header_compiler_direct":	buildpb.Attribute_LABEL_LIST,
	"heuristic_label_expansion":	buildpb.Attribute_BOOLEAN,
	"horizontal_resolution":	buildpb.Attribute_INTEGER,
	"host_platform":	buildpb.Attribute_BOOLEAN,
	"idl_import_root":	buildpb.Attribute_STRING,
	"idl_parcelables":	buildpb.Attribute_LABEL_LIST,
	"idl_preprocessed":	buildpb.Attribute_LABEL_LIST,
	"idl_srcs":	buildpb.Attribute_LABEL_LIST,
	"ijar":	buildpb.Attribute_LABEL_LIST,
	"import_prefix":	buildpb.Attribute_STRING,
	"imports":	buildpb.Attribute_STRING_LIST,
	"include_prefix":	buildpb.Attribute_STRING,
	"includes":	buildpb.Attribute_STRING_LIST,
	"incremental_dexing":	buildpb.Attribute_TRISTATE,
	"inline_constants":	buildpb.Attribute_BOOLEAN,
	"instruments":	buildpb.Attribute_LABEL,
	"interface_library":	buildpb.Attribute_LABEL,
	"interpreter":	buildpb.Attribute_LABEL,
	"interpreter_path":	buildpb.Attribute_STRING,
	"jacocorunner":	buildpb.Attribute_LABEL,
	"jars":	buildpb.Attribute_LABEL_LIST,
	"java":	buildpb.Attribute_LABEL,
	"java_home":	buildpb.Attribute_STRING,
	"javabuilder":	buildpb.Attribute_LABEL_LIST,
	"javabuilder_jvm_opts":	buildpb.Attribute_STRING_LIST,
	"javac":	buildpb.Attribute_LABEL_LIST,
	"javac_supports_multiplex_workers":	buildpb.Attribute_BOOLEAN,
	"javac_supports_workers":	buildpb.Attribute_BOOLEAN,
	"javacopts":	buildpb.Attribute_STRING_LIST,
	"jre_deps":	buildpb.Attribute_LABEL_LIST,
	"jvm_flags":	buildpb.Attribute_STRING_LIST,
	"jvm_opts":	buildpb.Attribute_STRING_LIST,
	"launcher":	buildpb.Attribute_LABEL,
	"launcher_uses_whole_archive":	buildpb.Attribute_BOOLEAN,
	"legacy_create_init":	buildpb.Attribute_TRISTATE,
	"libc_top":	buildpb.Attribute_LABEL,
	"licenses":	buildpb.Attribute_LICENSE,
	"linker_files":	buildpb.Attribute_LABEL,
	"linkopts":	buildpb.Attribute_STRING_LIST,
	"linkshared":	buildpb.Attribute_BOOLEAN,
	"linkstamp":	buildpb.Attribute_LABEL,
	"linkstatic":	buildpb.Attribute_BOOLEAN,
	"local":	buildpb.Attribute_BOOLEAN,
	"local_defines":	buildpb.Attribute_STRING_LIST,
	"local_versions":	buildpb.Attribute_LABEL,
	"main":	buildpb.Attribute_LABEL,
	"main_class":	buildpb.Attribute_STRING,
	"main_dex_classes":	buildpb.Attribute_LABEL,
	"main_dex_list":	buildpb.Attribute_LABEL,
	"main_dex_list_creator":	buildpb.Attribute_LABEL,
	"main_dex_list_opts":	buildpb.Attribute_STRING_LIST,
	"main_dex_proguard_specs":	buildpb.Attribute_LABEL_LIST,
	"malloc":	buildpb.Attribute_LABEL,
	"manifest":	buildpb.Attribute_LABEL,
	"manifest_values":	buildpb.Attribute_STRING_DICT,
	"message":	buildpb.Attribute_STRING,
	"minimum_os_version":	buildpb.Attribute_STRING,
	"misc":	buildpb.Attribute_STRING_LIST,
	"mnemonics":	buildpb.Attribute_STRING_LIST,
	"module_map":	buildpb.Attribute_LABEL,
	"module_name":	buildpb.Attribute_STRING,
	"multidex":	buildpb.Attribute_STRING,
	"name":	buildpb.Attribute_STRING,
	"neverlink":	buildpb.Attribute_BOOLEAN,
	"ninja_graph":	buildpb.Attribute_LABEL,
	"ninja_srcs":	buildpb.Attribute_LABEL_LIST,
	"nocompress_extensions":	buildpb.Attribute_STRING_LIST,
	"nocopts":	buildpb.Attribute_STRING,
	"non_arc_srcs":	buildpb.Attribute_LABEL_LIST,
	"objcopy_files":	buildpb.Attribute_LABEL,
	"oneversion":	buildpb.Attribute_LABEL,
	"oneversion_whitelist":	buildpb.Attribute_LABEL,
	"opts":	buildpb.Attribute_STRING_LIST,
	"os_constraints":	buildpb.Attribute_LABEL_LIST,
	"out":	buildpb.Attribute_STRING,
	"out_templates":	buildpb.Attribute_STRING_LIST,
	"output_group":	buildpb.Attribute_STRING,
	"output_groups":	buildpb.Attribute_STRING_LIST_DICT,
	"output_licenses":	buildpb.Attribute_LICENSE,
	"output_root":	buildpb.Attribute_STRING,
	"output_root_inputs":	buildpb.Attribute_STRING_LIST,
	"output_root_symlinks":	buildpb.Attribute_STRING_LIST,
	"output_to_bindir":	buildpb.Attribute_BOOLEAN,
	"outs":	buildpb.Attribute_STRING_LIST,
	"package_configuration":	buildpb.Attribute_LABEL_LIST,
	"package_metadata":	buildpb.Attribute_LABEL_LIST,
	"packages":	buildpb.Attribute_LABEL_LIST,
	"parents":	buildpb.Attribute_LABEL_LIST,
	"paropts":	buildpb.Attribute_STRING_LIST,
	"path":	buildpb.Attribute_STRING,
	"pch":	buildpb.Attribute_LABEL,
	"platform_apks":	buildpb.Attribute_LABEL_LIST,
	"platform_type":	buildpb.Attribute_STRING,
	"plugin":	buildpb.Attribute_LABEL,
	"plugins":	buildpb.Attribute_LABEL_LIST,
	"pregenerate_oat_files_for_tests":	buildpb.Attribute_BOOLEAN,
	"processor_class":	buildpb.Attribute_STRING,
	"profile":	buildpb.Attribute_LABEL,
	"proguard":	buildpb.Attribute_LABEL,
	"proguard_apply_dictionary":	buildpb.Attribute_LABEL,
	"proguard_apply_mapping":	buildpb.Attribute_LABEL,
	"proguard_generate_mapping":	buildpb.Attribute_BOOLEAN,
	"proguard_specs":	buildpb.Attribute_LABEL_LIST,
	"proto_profile":	buildpb.Attribute_LABEL,
	"provides_test_args":	buildpb.Attribute_BOOLEAN,
	"python_version":	buildpb.Attribute_STRING,
	"pytype_deps":	buildpb.Attribute_LABEL_LIST,
	"ram":	buildpb.Attribute_INTEGER,
	"reduced_classpath_incompatible_processors":	buildpb.Attribute_STRING_LIST,
	"reduced_classpath_incompatible_targets":	buildpb.Attribute_STRING_LIST,
	"reexport_deps":	buildpb.Attribute_LABEL_LIST,
	"remote_execution_properties":	buildpb.Attribute_STRING,
	"remote_versions":	buildpb.Attribute_LABEL,
	"repo_mapping":	buildpb.Attribute_STRING_DICT,
	"requires_action_output":	buildpb.Attribute_BOOLEAN,
	"resource_configuration_filters":	buildpb.Attribute_STRING_LIST,
	"resource_files":	buildpb.Attribute_LABEL_LIST,
	"resource_jars":	buildpb.Attribute_LABEL_LIST,
	"resource_strip_prefix":	buildpb.Attribute_STRING,
	"resourcejar":	buildpb.Attribute_LABEL_LIST,
	"resources":	buildpb.Attribute_LABEL_LIST,
	"restricted_to":	buildpb.Attribute_LABEL_LIST,
	"runtime":	buildpb.Attribute_LABEL,
	"runtime_deps":	buildpb.Attribute_LABEL_LIST,
	"scope":	buildpb.Attribute_LABEL_LIST,
	"screen_density":	buildpb.Attribute_INTEGER,
	"script":	buildpb.Attribute_LABEL,
	"sdk_dylibs":	buildpb.Attribute_STRING_LIST,
	"sdk_frameworks":	buildpb.Attribute_STRING_LIST,
	"sdk_includes":	buildpb.Attribute_STRING_LIST,
	"service_names":	buildpb.Attribute_STRING_LIST,
	"shard_count":	buildpb.Attribute_INTEGER,
	"shared_library":	buildpb.Attribute_LABEL,
	"shrink_resources":	buildpb.Attribute_TRISTATE,
	"shrinked_android_jar":	buildpb.Attribute_LABEL,
	"singlejar":	buildpb.Attribute_LABEL_LIST,
	"size":	buildpb.Attribute_STRING,
	"source_properties":	buildpb.Attribute_LABEL,
	"source_version":	buildpb.Attribute_STRING,
	"srcjar":	buildpb.Attribute_LABEL,
	"srcs":	buildpb.Attribute_LABEL_LIST,
	"srcs_version":	buildpb.Attribute_STRING,
	"stamp":	buildpb.Attribute_TRISTATE,
	"static_library":	buildpb.Attribute_LABEL,
	"static_runtime_lib":	buildpb.Attribute_LABEL,
	"strict":	buildpb.Attribute_BOOLEAN,
	"strict_deps":	buildpb.Attribute_BOOLEAN,
	"strict_exit":	buildpb.Attribute_BOOLEAN,
	"strip":	buildpb.Attribute_BOOLEAN,
	"strip_files":	buildpb.Attribute_LABEL,
	"strip_import_prefix":	buildpb.Attribute_STRING,
	"strip_include_prefix":	buildpb.Attribute_STRING,
	"stub_shebang":	buildpb.Attribute_STRING,
	"support_apks":	buildpb.Attribute_LABEL_LIST,
	"supports_header_parsing":	buildpb.Attribute_BOOLEAN,
	"supports_param_files":	buildpb.Attribute_BOOLEAN,
	"system_image":	buildpb.Attribute_LABEL,
	"system_provided":	buildpb.Attribute_BOOLEAN,
	"tags":	buildpb.Attribute_STRING_LIST,
	"target_compatible_with":	buildpb.Attribute_LABEL_LIST,
	"target_device":	buildpb.Attribute_LABEL,
	"target_platform":	buildpb.Attribute_BOOLEAN,
	"target_version":	buildpb.Attribute_STRING,
	"test_app":	buildpb.Attribute_LABEL,
	"test_class":	buildpb.Attribute_STRING,
	"testonly":	buildpb.Attribute_BOOLEAN,
	"tests":	buildpb.Attribute_LABEL_LIST,
	"textual_hdrs":	buildpb.Attribute_LABEL_LIST,
	"timeout":	buildpb.Attribute_STRING,
	"timezone_data":	buildpb.Attribute_LABEL,
	"toolchain":	buildpb.Attribute_STRING,
	"toolchain_config":	buildpb.Attribute_LABEL,
	"toolchain_identifier":	buildpb.Attribute_STRING,
	"toolchain_type":	buildpb.Attribute_LABEL,
	"toolchains":	buildpb.Attribute_LABEL_LIST,
	"tools":	buildpb.Attribute_LABEL_LIST,
	"transitive_configs":	buildpb.Attribute_STRING_LIST,
	"turbine_incompatible_processors":	buildpb.Attribute_STRING_LIST,
	"turbine_jvm_opts":	buildpb.Attribute_STRING_LIST,
	"use_testrunner":	buildpb.Attribute_BOOLEAN,
	"values":	buildpb.Attribute_STRING_DICT,
	"version":	buildpb.Attribute_STRING,
	"versions":	buildpb.Attribute_LABEL_LIST,
	"vertical_resolution":	buildpb.Attribute_INTEGER,
	"visibility":	buildpb.Attribute_STRING_LIST,
	"vm_heap":	buildpb.Attribute_INTEGER,
	"weak_sdk_frameworks":	buildpb.Attribute_STRING_LIST,
	"win_def_file":	buildpb.Attribute_LABEL,
	"working_directory":	buildpb.Attribute_STRING,
	"workspace_file":	buildpb.Attribute_STRING,
	"workspace_file_content":	buildpb.Attribute_STRING,
	"xlint":	buildpb.Attribute_STRING_LIST,
	"zipalign":	buildpb.Attribute_LABEL,
}
