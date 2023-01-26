# Discord-Happy-Birthday

## What Does It Do?

Well it of course wishes someone a happy birthday in Discord!

## Reason For Creation

It was my brother's birthday and I wanted to make something a bit whacky. I had it set up to trigger every time he spoke in the Discord server using his ID. 

There are two options to use. One is to use the command, the other is to use their Discord ID. 

- If the Discord ID is left empty then the command will be used. 

- If the Discord ID has a value it will be used.

## How to set up the script

In the script you will find 3 variables to change and 3 optional variables

1 (Required): Name
- Put in the name of the birthday holder

2 (Required): Age
- Put in the age of the birthday holder. This is how many times it will repeat the message. Giving them a birthday wish for every year they've been alive.

3 (Required): Birthday
- The month and day of the birthday. Used so the bot only works on their birthday.

4 (Optional): Pause
- How long you want the discord bot to wait in between birthday messages. Discord will slow you down though because it's basically spam, but funny spam.

5 (Optional): BirthdayID

- The Discord ID of the user whose birthday it is. You can get this in Discord in developer mode.

6 (Optional): Command

- This is the actual command used to wish them a happy birthday. You can change this to something else if you want.

## Options needed in the Discord Bot Developer Portal

- In the bot section make sure that "Message Content Intent" is enabled
- Bot permissions that are needed are:
    - Send Messages
    - Read Message History

## How to run
- cd /your/folder
- go run Happy_Birthday.go