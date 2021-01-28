/**
 * @fileoverview NodeJS binding to run the buildozer command through a convenient API
 */
import {spawnSync} from 'child_process';
// Use require() here since the file doesn't exist at design-time;
// bazel will copy the /launcher.js in this repo while building the package
const {getNativeBinary} = require('./buildozer');

/**
 * Models the data structure of a buildozer command file.
 * You should group as many commands as possible into one call to `run` for efficiency.
 * https://github.com/bazelbuild/buildtools/tree/master/buildozer#do-multiple-changes-at-once
 */
export interface CommandBatch {
    /**
     * Each entry is a buildozer edit command like 'new cc_library foo' or a print command
     * @see https://github.com/bazelbuild/buildtools/tree/master/buildozer#edit-commands
     */
    commands: string[];
    /**
     * Each entry is like a Bazel label
     * @see https://github.com/bazelbuild/buildtools/tree/master/buildozer#targets
     */
    targets: string[];
};

/**
 * run buildozer with a list of commands
 * @see https://github.com/bazelbuild/buildtools/tree/master/buildozer#do-multiple-changes-at-once
 * @returns The standard out of the buildozer command, split by lines
 */
export function run(...commands: CommandBatch[]): string[] {
    return runWithOptions(commands, {});
}

/**
 * run buildozer with a list of commands
 * @param commands a list of CommandBatch to pass to buildozer
 * @param options Options to pass to spawn
 * @param flags any buildozer flags to pass
 */
export function runWithOptions(commands: CommandBatch[], options: {cwd?: string}, flags: string[] = []): string[] {
    // From https://github.com/bazelbuild/buildtools/tree/master/buildozer#usage:
    // Here, label-list is a space-separated list of Bazel labels,
    // for example //path/to/pkg1:rule1 //path/to/pkg2:rule2.
    // Buildozer reads commands from FILE (- for stdin 
    //   (format: |-separated command line arguments to buildozer, excluding flags))
    const input = commands.map(c => [...c.commands, c.targets.join(',')].join('|')).join('\n');
    const {stdout, stderr, status, error} = spawnSync(getNativeBinary(), flags.concat([
        '-f', '-' /* read commands from stdin */
    ]), {
        ...options,
        input,
        encoding: 'utf-8',
    });
    // https://github.com/bazelbuild/buildtools/tree/master/buildozer#error-code
    if (status == 0 || status == 3) return stdout.trim().split('\n').filter(l => !!l);

    console.error(`buildozer exited with status ${status}\n${stderr}`);
    throw error;
}
