# Reddit Code Challenge
This project came about due as a response to an earlier code challenge I had completed.  I wanted to take the same requirements and implement them in go rather than C# and see the differences in the final result.
The original requirements are [here](Requirements.md).

## Reddit Tracker
This application sets up a monitor on a given subreddit to track which users post the most, and which posts have the highest total number of upvotes.

There are a couple of things to note:
- This only tracks data from the time the application starts until it stops running.  It does not look at historical data (except for votes on existing posts).
- The data is not persisted.  It is logged to the console.
- The user must provide their own AppId, AppSecret and RefreshToken (see below)
- The application returns the posts with the highest number of upvotes, not the highest netvote.

### Getting Started

Copy the `sample.env` file to `.env`.  Follow the instructions in the file to edit.  Add your secrets here.  This file is ignored by git and will not be checked in.

You can build and run locally using any standard IDE or the application can be run in a Docker container.  To quickly get up and running with a container, simply call the `./run` script, which will set up the container, build and execute the application for you.

## The Exercise
This was an interesting and fun challenge.  I had to work with an unfamiliar API and get set up as dev in their system.  I also got to write some enterprise-quality code that was a bit entertaining along the way.

My running notes as I was working can be found [here](Worklog.md).

My finally summary is [here]Notes.md

I hope this served it's purpose well.