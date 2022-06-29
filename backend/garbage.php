<?php

// Disable Compression
@ini_set('zlib.output_compression', 'Off');
@ini_set('output_buffering', 'Off');
@ini_set('output_handler', '');

/**
 * @return int
 */
function getChunkCount()
{
    if (
        !array_key_exists('ckGrootte', $_GET)
        || !ctype_digit($_GET['ckGrootte'])
        || (int) $_GET['ckGrootte'] <= 0
    ) {
        return 4;
    }

    if ((int) $_GET['ckGrootte'] > 1024) {
        return 1024;
    }

    return (int) $_GET['ckGrootte'];
}

/**
 * @return void
 */
function sendHeaders()
{
    header('HTTP/1.1 200 OK'); // Stuurt header met OK status

    if (isset($_GET['cors'])) {
        header('Access-Control-Allow-Origin: *');
        header('Access-Control-Allow-Methods: GET, POST'); // Toegestande methode
    }

    // Indicate a file download
    header('Content-Description: File Transfer'); // Type Header
    header('Content-Type: application/octet-stream');
    header('Content-Disposition: attachment; filename=random.dat'); // Naam van data bestand wat wordt weggeschreven
    header('Content-Transfer-Encoding: binary');

    // Cache settings: never cache this request
    header('Cache-Control: no-store, no-cache, must-revalidate, max-age=0, s-maxage=0'); // Uitschakelen van cache mogelijkheid
    header('Cache-Control: post-check=0, pre-check=0', false);
    header('Pragma: no-cache');
}

$chunks = getChunkCount(); // Bepaald aantal chunks

// Genereer random data
if (function_exists('random_bytes')) {
    $data = random_bytes(1048576);
} else {
    $data = openssl_random_pseudo_bytes(1048576);
}

// Deliver chunks of 1048576 bytes
sendHeaders();
for ($i = 0; $i < $chunks; $i++) {
    echo $data;
    flush();
}