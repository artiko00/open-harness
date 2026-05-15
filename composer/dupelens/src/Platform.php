<?php

declare(strict_types=1);

namespace OpenHarness\Dupelens;

/**
 * Detects the current OS/arch and returns the matching asset name from
 * the GitHub release. Mirrors npm's optional-deps platform tags.
 */
final class Platform
{
    /** @return array{os: string, arch: string, ext: string, asset: string} */
    public static function detect(string $tool): array
    {
        $os = self::os();
        $arch = self::arch();
        $ext = $os === 'windows' ? '.zip' : '.tar.gz';
        return [
            'os'    => $os,
            'arch'  => $arch,
            'ext'   => $ext,
            'asset' => "open-harness-{$tool}-{$os}-{$arch}{$ext}",
        ];
    }

    private static function os(): string
    {
        $family = PHP_OS_FAMILY;
        if ($family === 'Linux')   return 'linux';
        if ($family === 'Darwin')  return 'darwin';
        if ($family === 'Windows') return 'windows';
        throw new \RuntimeException(
            "open-harness: unsupported OS family '{$family}'. "
            . "Supported: Linux, Darwin (macOS), Windows."
        );
    }

    private static function arch(): string
    {
        $machine = strtolower(php_uname('m'));
        if (in_array($machine, ['x86_64', 'amd64'], true))  return 'x64';
        if (in_array($machine, ['arm64',  'aarch64'], true)) return 'arm64';
        throw new \RuntimeException(
            "open-harness: unsupported CPU arch '{$machine}'. "
            . "Supported: x86_64/amd64, arm64/aarch64."
        );
    }
}
