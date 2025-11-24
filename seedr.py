#!/usr/bin/env python3
# this script lets you interact with seedr.cc from the command line
# can be packaged into single executable using pyinstaller --onefile seedr.py

import argparse
import subprocess
import time
import os
from seedrcc import Login
from seedrcc import Seedr
from termcolor import colored
import humanize
from treelib import Node, Tree


def human_readable_bytes(byte_count):
    return humanize.naturalsize(byte_count, binary=True)

def getSeedrSettings(data):
    account = data['account']
    
    print(colored(f"Username: {account['username']}", "yellow"))
    print(colored(f"User ID: {account['user_id']}", "cyan"))
    
    
    space_used = human_readable_bytes(account['space_used'])
    space_max = human_readable_bytes(account['space_max'])
    bandwidth_used = human_readable_bytes(account['bandwidth_used'])
    
    print(colored(f"Space Used: {space_used}", "green"))
    print(colored(f"Space Max: {space_max}", "green"))
    print(colored(f"Bandwidth Used: {bandwidth_used}", "blue"))
    print(colored(f"Country: {data['country']}", "magenta"))


def fetchSeedrAccessToken():
    seedrFolder = os.getenv("HOME") + "/.cache/seedr"
    if not os.path.exists(seedrFolder):
        os.makedirs(seedrFolder)
    tokenLocation = seedrFolder + "/token.txt" 
    if not os.path.exists(tokenLocation):
        seedr = Login()

        device_code = seedr.getDeviceCode()
        print(colored(f"Device Code: {device_code}", "yellow"))

        time.sleep(30)
        response = seedr.authorize(device_code['device_code'])
        print(colored(f"Authorization Response: {response}", "green"))

        print(colored(f"Token: {seedr.token}", "cyan"))

        with open(tokenLocation, 'w') as f:
            f.write(seedr.token)
    else:
        token = open(tokenLocation, 'r').read().strip()
        def after_refresh(token):
            with open(tokenLocation, 'w') as f:
                f.write(token)
        Seedr(token, callbackFunc=after_refresh)
    
    return token

SeedrToken = fetchSeedrAccessToken()
account = Seedr(token=SeedrToken)

def addMagnet(magnet):
    try:
        response = account.addTorrent(magnet)
        result = response["result"]
        if result == "not_enough_space_added_to_wishlist":
            print(colored("Not enough space to add the torrent.", "red"))
        elif result == True:
            try:
                name = response["title"]
                print(colored(f"Added {name}.", "green"))
            except Exception as e:
                print(colored("Torrent already added.", "red"))
    except Exception as e:
        print("error: unable to send requests to seedr, try again later.")


def printSettings(): 
    settings = account.getSettings()
    getSeedrSettings(settings)

def listTorrentFolders(): 
    tree = Tree()
    data = account.listContents()
    username = account.getSettings()["account"]["username"]
    tree.create_node(f"/{username}", data["folder_id"])
    folders = data["folders"]
    for folder in folders: 
        torrentFolder = account.listContents(folder['id'])
        if not tree.get_node(torrentFolder['folder_id']):
            tree.create_node(f"{torrentFolder['fullname']} - {torrentFolder['folder_id']}", torrentFolder['folder_id'], parent=data["folder_id"])
        subfoldersIDs = torrentFolder["folders"]
        if subfoldersIDs:
            for subfolderID in subfoldersIDs:
                subFolder = account.listContents(subfolderID['id']) 
                if not tree.get_node(subFolder['folder_id']):
                    tree.create_node(f"{subFolder['name']}", subFolder['folder_id'], parent=subFolder['parent'])
                subFolderfiles = subFolder["files"]
                for file in subFolderfiles:
                    if not tree.get_node(file['folder_file_id']):
                        filesize = human_readable_bytes(file['size'])
                        tree.create_node(f"{file['name']} (ID: {file['folder_file_id']}) - (Size: {filesize})", file['folder_file_id'], parent=file["folder_id"]) 
            files = torrentFolder["files"]
            for file in files:
                if not tree.get_node(file['folder_file_id']):
                    filesize = human_readable_bytes(file['size'])
                    tree.create_node(f"{file['name']} (ID: {file['folder_file_id']}) - (Size: {filesize})", file['folder_file_id'], parent=file["folder_id"]) 
    print(tree.show(stdout=False)) 

def getDownloadUrl(isDirectory, FileId):
    url:str = None
    if isDirectory == False:
        file = account.fetchFile(FileId)
        name = file["name"] 
        url = file["url"]
        print(colored(f"Name: {name}", "blue"))
        print(colored(f"URL: {url}", "magenta"))
    elif isDirectory == True:
        directory = account.createArchive(FileId)
        url = directory["archive_url"]
        print(colored(f"URL: {url}", "magenta"))
    try:
        subprocess.run(['wl-copy', url])
    except Exception as e:
        print("error: while copying url to clipboard, make sure wl-clipboard is installed.")


def deleteTorrentFolder(FolderID):
    response = account.deleteFolder(FolderID)
    if response['result'] == True:
        print(colored("successfully deleted the folder", "blue"))
    else:
        print(colored("unable to delete the folder.", "red"))


def main():
    parser = argparse.ArgumentParser(description="seedr") 
    parser.add_argument('--settings', action='store_true', help="get account settings")
    parser.add_argument('--login', action='store_true', help="log into seedr")
    parser.add_argument('--add', type=str, help="add torrents to seedr")
    parser.add_argument('--list', action='store_true', help="list folders on seedr") 
    parser.add_argument('--rm', type=str, help="delete a torrent folder")
    parser.add_argument('--get', action='store_true', help="get download url of files/folders")
    parser.add_argument('-f', '--file', type=str, help="specify a file path when using --get.")
    parser.add_argument('-d', '--directory', type=str, help="specify a directory path when using --get.")


    args = parser.parse_args()

    if args.settings:
        printSettings()


    if args.list:
        listTorrentFolders()
    
    if args.login:
        print(colored("ogging in and retrieving token...", "yellow"))
        fetchSeedrAccessToken()
        print(colored("ogin successful and token stored.", "green"))
    
    if args.add:
        addMagnet(args.add)

    if args.get:
        if args.file:
            getDownloadUrl(False, args.file)
        elif args.directory:
            getDownloadUrl(True, args.directory)

    if args.rm:
        deleteTorrentFolder(args.rm)

if __name__ == "__main__":
    main()
