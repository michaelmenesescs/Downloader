#!/usr/bin/env python

import re
import subprocess
import sys


def download_tidal(url, username, password):
    tidal_dl_command = ["tidal-dl", "-u", username, "-p", password, url]
    subprocess.run(tidal_dl_command)


def download_soundcloud(url):
    scdl_command = ["scdl -l", url]
    subprocess.run(scdl_command)


def download_youtube(url):
    yttomp3_command = ["ytmdl", url]
    subprocess.run(yttomp3_command)


def main():
    while True:
        print("Download media from Tidal, Soundcloud, or YouTube.")

        url = input("URL: ")

        if re.match(r"https?://(www\.)?soundcloud.com/", url):
            
            download_soundcloud(url)
        elif re.match(r"https?://listen.tidal.com/", url) or re.match(r"https?://tidal.com/", url) :
            print("Tidal")
            username = input("Tidal username: ")
            password = input("Tidal password: ")
            download_tidal(url, username, password)
        elif re.match(r"https?://(www\.)?youtube.com/", url) or re.match(r"https?://(www\.)?youtu.be/", url):
            print("YouTube")
            download_youtube(url)
        else:
            print(f"Error: URL {url} is not supported.", file=sys.stderr)
            sys.exit(1)


if __name__ == "__main__":
    main()
