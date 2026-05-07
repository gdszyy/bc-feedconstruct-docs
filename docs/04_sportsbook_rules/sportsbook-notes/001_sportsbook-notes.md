---
title: SportsBook Notes
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=sportsBookNotes
current_loc: sportsBookNotes
location: root
top_category: SPORTSBOOK NOTES
product_line: 体育博彩业务规则
business_domain: 体育博彩业务规则
scraped_at: 2026-05-07T08:49:13.195Z
---

# SportsBook Notes

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=sportsBookNotes`。

| 字段 | 值 |
|---|---|
| 一级分类 | SPORTSBOOK NOTES |
| 产品线 | 体育博彩业务规则 |
| 业务域 | 体育博彩业务规则 |
| currentLoc | `sportsBookNotes` |
| location | `root` |

## 文档正文
SportsBook Notes

- Live Delay Logic

To prevent late bet acceptance - when an event (e.g., goal, corner shot, point, etc) occurs before the market is technically closed - Live Delay (LD) is applied. This introduces a controlled delay window to ensure that bets are not accepted on outcomes that have already happened but are not yet reflected in the market status.

LD Field Hierarchy Each object type has an LD field. The final value used is resolved by applying the following bottom-up priority logic:

Match LD = Competition.LD (if set) → otherwise Sport.LD (Always a positive value)

Selection LD = SelectionType.LD (if set) → otherwise MarketType.LD (Can be positive or negative)

Bet LD = Match LD + Selection LD

**Note:** The Live Delay field in the Match Update command (where Type: Match) contains a pre-calculated value for the live delay at the time of the update. Consequently, Partners do not need to perform additional calculations to determine the exact Live Delay at the Match level.

- The Live Delay field is only sent during the **Live offer** phase of a match.
- SelectionType Live Delay is included in getSelectionTypes and getSelectionTypeByID commands, provided a value for Live Delay has been assigned to that selection type.

- **Calculating Bet Live Delay**:

  To determine the Live Delay for a specific Bet, Partners must use the following formula:

  Bet LD = Match LD + Selection LD

  Example: If the Match LD is **2** and the Selection LD is **0**, the resulting Bet Live Delay will be **2 seconds**.

- **Use Case: Preventing Late Bets Example scenario**:

  A goal occurred at 13:18:45

  The market was closed at 13:18:47

  Match LD is set to 5 seconds

  Selection LD is 0 → so, Bet LD = 5 (the information is provided with the `LiveDelay` field in `Command`:`MatchUpdate`, `Type`:`Match` updates).

  Even though the market was closed 2 seconds after the event, the Live Delay logic ensures that any bets placed between 13:18:45 and 13:18:47 are blocked. This is because, on our end, the system waits for 5 seconds (as defined by Bet LD) before accepting bets, effectively preventing late bets from being accepted during short closure delays.

- **Partner Integration Notes**:

  For **OddsFeed-only** partners, using the LiveDelay field is optional. We provide this field purely as informational, to reflect how we handle bets on our side.

  For **OddsFeed + BetGuard** partners, the LiveDelay is applied as sent in the OddsFeed updates. Since BetGuard handles bet acceptance on our side, the system fully enforces the LiveDelay logic to decide whether a received bet should be accepted or rejected.

- Matches with Parent ID

In some cases, some markets from a match are offered under a separate match, for example, the Cards and Corners related markets of a World Cup match in Live are offered separately from the main event under a Competition called “World Cup. Cards and Corners”. In order to receive these markets for Live the ‘side’ match (or the competition) should be booked as well. The Parent IDs of the `side` matches is the Match ID of the main match, so if you don`t want to offer them separately, you can merge the matches by the Partner ID. Please note, that for invoicing purposes the matches which have a Parent ID are not counted, which means that if you have the main and `side` matches booked, they will be counted as 1 match.

- Void Notifications

- **Void Notification**

  There may be situations when the selections are resulted, but there was an inaccuracy, such as a wrong odd offered or a wrong score; in these cases, our traders Void the previous result and send it with the Void Notification, along with all necessary information such as the Time Frame and the Reason for Voiding. Please keep in mind that the Time Frame in Void Notification is based on FeedConstruct bets, and because there is no way to monitor OddsFeed Partners bets, the Time Frame indicated there may not be the same as OddsFeed Partners have on their end.
- **Unvoid Notification**

  There is an Unvoid Notification, which is sent with all relevant information, such as the Time Frame and Reason for Unvoiding, when an already Voided selection is Unvoided. Please keep in mind that the Time Frame in Void Notification is based on FeedConstruct bets, and because there is no way to monitor OddsFeed Partners bets, the Time Frame indicated there may not be the same as OddsFeed Partners have on their end.

  Both of them are specifically intended as Notifications, and OddsFeed Partners are free to ignore them.

- Removing a match from PreMatch

In case a Match StartDate is reached and there was no update for Match Start (check for the fields and values [here](../documentation?currentLoc=match_lifecycle_for_live)), the Match should be removed from the PreMatch offer and be opened it in the Live offer after the Start update.

- IsTeamReversed

The isTeamReversed flag indicates that in some North American competitions (like NBA, NFL), the away team is considered the home team as a gesture of honour, reflected on the official website and in the data feed. This means the away team is sent as the home team according to the official source. The flag is included with competition updates and covers the whole competition.

If isTeamReversed is true, it means the home team is not the owner of the stadium, and the partner has the option to flip the data (teams' positions, scores, market data) to show the names in the classical way if preferred.

If you do not want to offer the names reversed as the competition does, you can use the isTeamReversed flag to detect the situation and then flip the data back to the classical order yourself. This involves switching the positions of teams, their scores, and any related market information to reflect the traditional home/away team naming.
