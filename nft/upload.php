<?php

// Change The Secret Key to the one used in the Go Program
define('UPLOAD_SECRET_KEY', '630d96ae6fe44400566a781dfcb7f222');

// By default the files will be uploaded to the same folder where the script runs. Change it if necessary
define('NFT_FOLDER', dirname(__FILE__));


// Do not change the lines below
if (!isset($_POST['secret_key'])) {
  echo "No Secret Key Sent";
  return;
}

if ($_POST['secret_key'] != UPLOAD_SECRET_KEY) {
  echo "Incorrect secret key";
  return;
}

if (isset($_FILES['nft'])) {
  if (!saveFile($_FILES['nft']))
    return;
}

if (isset($_FILES['description'])) {
  if (!saveFile($_FILES['description']))
    return;
}

if (isset($_FILES['unique'])) {
  if (!saveFile($_FILES['unique'])) 
    return;
}

function saveFile($item) {
  if ($item['error'] != ERR_UPLOAD_OK) {
    echo "Failed to upload " . $item['name'] . ". Error {$item['error']}";
    return false;
  }

  $uploadFile = NFT_FOLDER . "/" . basename($item['name']);
  if (!move_uploaded_file($item['tmp_name'], $uploadFile)) {
    echo "Failed to move uploaded file {$item['name']}";
    return false;
  }

  return true;

}

echo "OK";

?>


