# Use the official Python image as the base image
FROM python:3

# Set the working directory to /app
WORKDIR /app

# Install the required dependencies
RUN apt-get update -y 

RUN apt-get install -y git

RUN apt-get install -y ffmpeg

# Clone the Tidal Media Downloader repository
RUN git clone https://github.com/yaronzz/Tidal-Media-Downloader.git

# Install Tidal Media Downloader
RUN pip install -r Tidal-Media-Downloader/TIDALDL-PY/requirements.txt
RUN python Tidal-Media-Downloader/TIDALDL-PY/setup.py install

# Install SCDL
RUN pip3 install scdl

# Clone the YTtomp3 repository
RUN pip install ytmdl --upgrade

# Add the script that checks the URL and runs the appropriate downloader
COPY download.py /app/download.py
RUN chmod +x /app/download.py

# Define the volume
VOLUME /app/downloads

# Set the entrypoint to the script that checks the URL and runs the appropriate downloader
ENTRYPOINT ["/app/download.py"]
