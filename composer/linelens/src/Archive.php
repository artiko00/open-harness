<?php

declare(strict_types=1);

namespace OpenHarness\Linelens;

/**
 * Minimal archive extractor: pulls a single named binary out of a
 * tarball (Linux/macOS) or zip (Windows) without writing the rest of
 * the archive to disk.
 */
final class Archive
{
    public static function extractBinary(
        string $blob,
        string $ext,
        string $name,
        string $dest
    ): void {
        $needle = $ext === '.zip' ? $name . '.exe' : $name;
        $tmp = tempnam(sys_get_temp_dir(), 'oh-archive-');
        if ($tmp === false || file_put_contents($tmp, $blob) === false) {
            throw new \RuntimeException("cannot stage archive at {$tmp}");
        }
        try {
            $ext === '.zip'
                ? self::extractZip($tmp, $needle, $dest)
                : self::extractTarGz($tmp, $needle, $dest);
        } finally {
            @unlink($tmp);
        }
    }

    private static function extractZip(string $path, string $entry, string $dest): void
    {
        $zip = new \ZipArchive();
        if ($zip->open($path) !== true) {
            throw new \RuntimeException("cannot open zip {$path}");
        }
        try {
            $data = $zip->getFromName($entry);
            if ($data === false) {
                throw new \RuntimeException("zip missing entry {$entry}");
            }
            if (file_put_contents($dest, $data) === false) {
                throw new \RuntimeException("cannot write {$dest}");
            }
        } finally {
            $zip->close();
        }
    }

    private static function extractTarGz(string $path, string $entry, string $dest): void
    {
        $phar = new \PharData($path);
        $stream = $phar->offsetGet($entry);
        if (!$stream instanceof \PharFileInfo) {
            throw new \RuntimeException("tarball missing entry {$entry}");
        }
        $data = file_get_contents($stream->getPathname());
        if ($data === false || file_put_contents($dest, $data) === false) {
            throw new \RuntimeException("cannot write {$dest}");
        }
    }
}
