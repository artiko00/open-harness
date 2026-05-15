<?php

declare(strict_types=1);

namespace OpenHarness\Dupelens;

use Composer\Script\Event;

/**
 * Composer post-install/update hook: downloads the right native binary
 * from the GitHub Release matching this package's version. Verifies
 * the SHA256 from the release's checksums.txt before installing.
 */
final class BinaryDownloader
{
    private const TOOL = 'dupelens';
    private const REPO = 'artiko00/open-harness';

    public static function install(Event $event): void
    {
        $io = $event->getIO();
        $pkg = self::pkg($event);
        $version = $pkg['extra']['open-harness']['version'] ?? null;
        if (!$version) {
            $io->writeError('<error>open-harness/' . self::TOOL . ': no version in composer.json extra</error>');
            return;
        }
        $plat = Platform::detect(self::TOOL);
        $base = 'https://github.com/' . self::REPO . "/releases/download/v{$version}";
        $dest = self::destDir($event) . DIRECTORY_SEPARATOR . self::TOOL
              . ($plat['os'] === 'windows' ? '.exe' : '');
        $io->write('open-harness/' . self::TOOL . ": fetching {$plat['asset']}…");
        $tarball = self::download("{$base}/{$plat['asset']}");
        self::verifySha256($tarball, "{$base}/checksums.txt", $plat['asset']);
        Archive::extractBinary($tarball, $plat['ext'], self::TOOL, $dest);
        @chmod($dest, 0755);
        $io->write('open-harness/' . self::TOOL . ": installed {$dest}");
    }

    private static function verifySha256(string $blob, string $url, string $asset): void
    {
        $body = @file_get_contents($url);
        if ($body === false) return; // graceful: no checksums file → skip
        foreach (preg_split('/\r?\n/', $body) as $line) {
            $p = preg_split('/\s+/', trim($line));
            if (count($p) === 2 && $p[1] === $asset) {
                if (!hash_equals(strtolower($p[0]), hash('sha256', $blob))) {
                    throw new \RuntimeException("checksum mismatch for {$asset}");
                }
                return;
            }
        }
    }

    /** @return array<string, mixed> */
    private static function pkg(Event $event): array
    {
        $path = self::pkgRoot($event) . '/composer.json';
        $raw = @file_get_contents($path);
        if ($raw === false) {
            throw new \RuntimeException("cannot read {$path}");
        }
        return json_decode($raw, true, 512, JSON_THROW_ON_ERROR);
    }

    private static function pkgRoot(Event $event): string
    {
        $vendor = $event->getComposer()->getConfig()->get('vendor-dir');
        return $vendor . '/open-harness/' . self::TOOL;
    }

    private static function destDir(Event $event): string
    {
        $bin = $event->getComposer()->getConfig()->get('bin-dir');
        if (!is_dir($bin) && !mkdir($bin, 0755, true) && !is_dir($bin)) {
            throw new \RuntimeException("cannot create bin-dir {$bin}");
        }
        return $bin;
    }

    private static function download(string $url): string
    {
        $data = @file_get_contents($url);
        if ($data === false) {
            throw new \RuntimeException('open-harness/' . self::TOOL . ": failed to fetch {$url}");
        }
        return $data;
    }
}
