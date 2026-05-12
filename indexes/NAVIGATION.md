# BC FeedConstruct 文档导航索引

本索引由 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=summary` 的公开访客文档抓取生成。原始抓取页面数为 **261**，按 `currentLoc + location` 去重后的 Markdown 页面数为 **249**。

## 本地规划文档（非外部抓取）

| 路径 | 用途 |
|---|---|
| [`docs/07_frontend_architecture/README.md`](../docs/07_frontend_architecture/README.md) | 体育博彩前端数据来源整体方案：模块（M01~M16）、页面（P01~P08）骨架、Go BFF↔Next.js 契约、状态机、验收清单。 |

## 按一级文档模块索引

### RABBIT MQ (`oddsFeedRmqAndWebApi`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Summary | `summary` | 数据源服务 / OddsFeed | [001_summary.md](../docs/01_data_feed/rmq-web-api/001_summary.md) |
| 2 | Access method | `access` | 数据源服务 / OddsFeed | [002_access.md](../docs/01_data_feed/rmq-web-api/002_access.md) |
| 3 | Integration Guide | `integrationGuide` | 数据源服务 / OddsFeed | [003_integrationguide.md](../docs/01_data_feed/rmq-web-api/003_integrationguide.md) |
| 4 | Objects Descriptions | `objectsDescriptions` | 数据源服务 / OddsFeed | [004_objectsdescriptions.md](../docs/01_data_feed/rmq-web-api/004_objectsdescriptions.md) |
| 5 | Sport | `sport` | 数据源服务 / OddsFeed | [005_sport.md](../docs/01_data_feed/rmq-web-api/005_sport.md) |
| 6 | Region | `region` | 数据源服务 / OddsFeed | [006_region.md](../docs/01_data_feed/rmq-web-api/006_region.md) |
| 7 | Competition | `competition` | 数据源服务 / OddsFeed | [007_competition.md](../docs/01_data_feed/rmq-web-api/007_competition.md) |
| 8 | Match | `match` | 数据源服务 / OddsFeed | [008_match.md](../docs/01_data_feed/rmq-web-api/008_match.md) |
| 9 | Market | `market` | 数据源服务 / OddsFeed | [009_market.md](../docs/01_data_feed/rmq-web-api/009_market.md) |
| 10 | Selection | `selection` | 数据源服务 / OddsFeed | [010_selection.md](../docs/01_data_feed/rmq-web-api/010_selection.md) |
| 11 | Stat | `stat` | 数据源服务 / OddsFeed | [011_stat.md](../docs/01_data_feed/rmq-web-api/011_stat.md) |
| 12 | MatchMember | `matchMember` | 数据源服务 / OddsFeed | [012_matchmember.md](../docs/01_data_feed/rmq-web-api/012_matchmember.md) |
| 13 | MarketType | `marketType` | 数据源服务 / OddsFeed | [013_markettype.md](../docs/01_data_feed/rmq-web-api/013_markettype.md) |
| 14 | SelectionType | `selectionType` | 数据源服务 / OddsFeed | [014_selectiontype.md](../docs/01_data_feed/rmq-web-api/014_selectiontype.md) |
| 15 | VoidNotification | `voidNotification` | 数据源服务 / OddsFeed | [015_voidnotification.md](../docs/01_data_feed/rmq-web-api/015_voidnotification.md) |
| 16 | PartnerBooking | `partnerBooking` | 数据源服务 / OddsFeed | [016_partnerbooking.md](../docs/01_data_feed/rmq-web-api/016_partnerbooking.md) |
| 17 | Calendar Match | `calendar-match` | 数据源服务 / OddsFeed | [017_calendar-match.md](../docs/01_data_feed/rmq-web-api/017_calendar-match.md) |
| 18 | Book | `book` | 数据源服务 / OddsFeed | [018_book.md](../docs/01_data_feed/rmq-web-api/018_book.md) |
| 19 | EventType | `event-types` | 数据源服务 / OddsFeed | [019_event-types.md](../docs/01_data_feed/rmq-web-api/019_event-types.md) |
| 20 | Period | `periods` | 数据源服务 / OddsFeed | [020_periods.md](../docs/01_data_feed/rmq-web-api/020_periods.md) |
| 21 | UnBookObject | `unBookObject` | 数据源服务 / OddsFeed | [021_unbookobject.md](../docs/01_data_feed/rmq-web-api/021_unbookobject.md) |
| 22 | BookObject | `bookObject` | 数据源服务 / OddsFeed | [022_bookobject.md](../docs/01_data_feed/rmq-web-api/022_bookobject.md) |
| 23 | MarketTypeBooking | `marketTypeBooking` | 数据源服务 / OddsFeed | [023_markettypebooking.md](../docs/01_data_feed/rmq-web-api/023_markettypebooking.md) |
| 24 | RacingInfo | `racingInfo` | 数据源服务 / OddsFeed | [024_racinginfo.md](../docs/01_data_feed/rmq-web-api/024_racinginfo.md) |
| 25 | MarketExtraInfo | `marketExtraInfo` | 数据源服务 / OddsFeed | [025_marketextrainfo.md](../docs/01_data_feed/rmq-web-api/025_marketextrainfo.md) |
| 26 | BetDivident | `betDivident` | 数据源服务 / OddsFeed | [026_betdivident.md](../docs/01_data_feed/rmq-web-api/026_betdivident.md) |
| 27 | ManualEachWayTerms | `manualEachWayTerms` | 数据源服务 / OddsFeed | [027_manualeachwayterms.md](../docs/01_data_feed/rmq-web-api/027_manualeachwayterms.md) |
| 28 | HorseInfo | `horseInfo` | 数据源服务 / OddsFeed | [028_horseinfo.md](../docs/01_data_feed/rmq-web-api/028_horseinfo.md) |
| 29 | SportOrder | `sportOrder` | 数据源服务 / OddsFeed | [029_sportorder.md](../docs/01_data_feed/rmq-web-api/029_sportorder.md) |
| 30 | CommandResponse | `commandResponse` | 数据源服务 / OddsFeed | [030_commandresponse.md](../docs/01_data_feed/rmq-web-api/030_commandresponse.md) |
| 31 | ResponseError | `responseError` | 数据源服务 / OddsFeed | [031_responseerror.md](../docs/01_data_feed/rmq-web-api/031_responseerror.md) |
| 32 | Message Syntax | `messageSyntax` | 数据源服务 / OddsFeed | [032_messagesyntax.md](../docs/01_data_feed/rmq-web-api/032_messagesyntax.md) |
| 33 | Web Methods | `webMethods` | 数据源服务 / OddsFeed | [033_webmethods.md](../docs/01_data_feed/rmq-web-api/033_webmethods.md) |
| 34 | Objects as part of the feed | `objectsAsPartFeed` | 数据源服务 / OddsFeed | [034_objectsaspartfeed.md](../docs/01_data_feed/rmq-web-api/034_objectsaspartfeed.md) |
| 35 | Object Types | `objectTypes` | 数据源服务 / OddsFeed | [035_objecttypes.md](../docs/01_data_feed/rmq-web-api/035_objecttypes.md) |
| 36 | Integration Notes | `integrationNotes` | 数据源服务 / OddsFeed | [036_integrationnotes.md](../docs/01_data_feed/rmq-web-api/036_integrationnotes.md) |
| 37 | Result Codes | `resultCodes` | 数据源服务 / OddsFeed | [037_resultcodes.md](../docs/01_data_feed/rmq-web-api/037_resultcodes.md) |
| 38 | Samples | `samples` | 数据源服务 / OddsFeed | [038_samples.md](../docs/01_data_feed/rmq-web-api/038_samples.md) |
| 39 | .Net/C# | `cSharp` | 数据源服务 / OddsFeed | [039_csharp.md](../docs/01_data_feed/rmq-web-api/039_csharp.md) |
| 40 | Node.js | `nodeJs` | 数据源服务 / OddsFeed | [040_nodejs.md](../docs/01_data_feed/rmq-web-api/040_nodejs.md) |
| 41 | PHP | `php` | 数据源服务 / OddsFeed | [041_php.md](../docs/01_data_feed/rmq-web-api/041_php.md) |
| 42 | Java | `java` | 数据源服务 / OddsFeed | [042_java.md](../docs/01_data_feed/rmq-web-api/042_java.md) |
| 43 | Change Log | `changeLog` | 数据源服务 / OddsFeed | [043_changelog.md](../docs/01_data_feed/rmq-web-api/043_changelog.md) |

### TCP SOCKET (`feedSocketApi`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Summary | `summary` | 数据源服务 / OddsFeed | [001_summary.md](../docs/01_data_feed/tcp-socket-api/001_summary.md) |
| 2 | Access Method | `access` | 数据源服务 / OddsFeed | [002_access.md](../docs/01_data_feed/tcp-socket-api/002_access.md) |
| 3 | Integration Guide | `integrationGuide` | 数据源服务 / OddsFeed | [003_integrationguide.md](../docs/01_data_feed/tcp-socket-api/003_integrationguide.md) |
| 4 | Objects Descriptions | `objectsDescriptions` | 数据源服务 / OddsFeed | [004_objectsdescriptions.md](../docs/01_data_feed/tcp-socket-api/004_objectsdescriptions.md) |
| 5 | Sport | `sport` | 数据源服务 / OddsFeed | [005_sport.md](../docs/01_data_feed/tcp-socket-api/005_sport.md) |
| 6 | Region | `region` | 数据源服务 / OddsFeed | [006_region.md](../docs/01_data_feed/tcp-socket-api/006_region.md) |
| 7 | Competition | `competition` | 数据源服务 / OddsFeed | [007_competition.md](../docs/01_data_feed/tcp-socket-api/007_competition.md) |
| 8 | Match | `match` | 数据源服务 / OddsFeed | [008_match.md](../docs/01_data_feed/tcp-socket-api/008_match.md) |
| 9 | Market | `market` | 数据源服务 / OddsFeed | [009_market.md](../docs/01_data_feed/tcp-socket-api/009_market.md) |
| 10 | Selection | `selection` | 数据源服务 / OddsFeed | [010_selection.md](../docs/01_data_feed/tcp-socket-api/010_selection.md) |
| 11 | Stat | `stat` | 数据源服务 / OddsFeed | [011_stat.md](../docs/01_data_feed/tcp-socket-api/011_stat.md) |
| 12 | MatchMember | `matchMember` | 数据源服务 / OddsFeed | [012_matchmember.md](../docs/01_data_feed/tcp-socket-api/012_matchmember.md) |
| 13 | MarketType | `marketType` | 数据源服务 / OddsFeed | [013_markettype.md](../docs/01_data_feed/tcp-socket-api/013_markettype.md) |
| 14 | SelectionType | `selectionType` | 数据源服务 / OddsFeed | [014_selectiontype.md](../docs/01_data_feed/tcp-socket-api/014_selectiontype.md) |
| 15 | VoidNotification | `voidNotification` | 数据源服务 / OddsFeed | [015_voidnotification.md](../docs/01_data_feed/tcp-socket-api/015_voidnotification.md) |
| 16 | PartnerBooking | `partnerBooking` | 数据源服务 / OddsFeed | [016_partnerbooking.md](../docs/01_data_feed/tcp-socket-api/016_partnerbooking.md) |
| 17 | Calendar Match | `calendar-match` | 数据源服务 / OddsFeed | [017_calendar-match.md](../docs/01_data_feed/tcp-socket-api/017_calendar-match.md) |
| 18 | BookItems | `book-items` | 数据源服务 / OddsFeed | [018_book-items.md](../docs/01_data_feed/tcp-socket-api/018_book-items.md) |
| 19 | EventType | `event-types` | 数据源服务 / OddsFeed | [019_event-types.md](../docs/01_data_feed/tcp-socket-api/019_event-types.md) |
| 20 | Period | `periods` | 数据源服务 / OddsFeed | [020_periods.md](../docs/01_data_feed/tcp-socket-api/020_periods.md) |
| 21 | UnBookObject | `unBookObject` | 数据源服务 / OddsFeed | [021_unbookobject.md](../docs/01_data_feed/tcp-socket-api/021_unbookobject.md) |
| 22 | BookObject | `bookObject` | 数据源服务 / OddsFeed | [022_bookobject.md](../docs/01_data_feed/tcp-socket-api/022_bookobject.md) |
| 23 | MarketTypeBooking | `marketTypeBooking` | 数据源服务 / OddsFeed | [023_markettypebooking.md](../docs/01_data_feed/tcp-socket-api/023_markettypebooking.md) |
| 24 | SportOrder | `sportOrder` | 数据源服务 / OddsFeed | [024_sportorder.md](../docs/01_data_feed/tcp-socket-api/024_sportorder.md) |
| 25 | ResponseError | `responseError` | 数据源服务 / OddsFeed | [025_responseerror.md](../docs/01_data_feed/tcp-socket-api/025_responseerror.md) |
| 26 | Message Syntax | `messageSyntax` | 数据源服务 / OddsFeed | [026_messagesyntax.md](../docs/01_data_feed/tcp-socket-api/026_messagesyntax.md) |
| 27 | Command List | `commandList` | 数据源服务 / OddsFeed | [027_commandlist.md](../docs/01_data_feed/tcp-socket-api/027_commandlist.md) |
| 28 | Objects as part of the feed | `objectsAsPartFeed` | 数据源服务 / OddsFeed | [028_objectsaspartfeed.md](../docs/01_data_feed/tcp-socket-api/028_objectsaspartfeed.md) |
| 29 | Object Types | `objectTypes` | 数据源服务 / OddsFeed | [029_objecttypes.md](../docs/01_data_feed/tcp-socket-api/029_objecttypes.md) |
| 30 | Integration Notes | `integrationNotes` | 数据源服务 / OddsFeed | [030_integrationnotes.md](../docs/01_data_feed/tcp-socket-api/030_integrationnotes.md) |
| 31 | Result Codes | `resultCodes` | 数据源服务 / OddsFeed | [031_resultcodes.md](../docs/01_data_feed/tcp-socket-api/031_resultcodes.md) |
| 32 | Samples | `samples` | 数据源服务 / OddsFeed | [032_samples.md](../docs/01_data_feed/tcp-socket-api/032_samples.md) |
| 33 | Change Log | `changeLog` | 数据源服务 / OddsFeed | [033_changelog.md](../docs/01_data_feed/tcp-socket-api/033_changelog.md) |

### TRANSLATIONS SOCKET API (`translationSocketApi`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Summary | `summary` | 翻译数据服务 | [001_summary.md](../docs/02_translations/translations-socket-api/001_summary.md) |
| 2 | Access Method | `access` | 翻译数据服务 | [002_access.md](../docs/02_translations/translations-socket-api/002_access.md) |
| 3 | Integration Guide | `integrationGuide` | 翻译数据服务 | [003_integrationguide.md](../docs/02_translations/translations-socket-api/003_integrationguide.md) |
| 4 | Objects Descriptions | `objectsDescriptions` | 翻译数据服务 | [004_objectsdescriptions.md](../docs/02_translations/translations-socket-api/004_objectsdescriptions.md) |
| 5 | Message Syntax | `messageSyntax` | 翻译数据服务 | [005_messagesyntax.md](../docs/02_translations/translations-socket-api/005_messagesyntax.md) |
| 6 | Command List | `commandList` | 翻译数据服务 | [006_commandlist.md](../docs/02_translations/translations-socket-api/006_commandlist.md) |
| 7 | Objects as part of the feed | `objectsAsPartFeed` | 翻译数据服务 | [007_objectsaspartfeed.md](../docs/02_translations/translations-socket-api/007_objectsaspartfeed.md) |
| 8 | Change Log | `changeLog` | 翻译数据服务 | [008_changelog.md](../docs/02_translations/translations-socket-api/008_changelog.md) |

### TRANSLATIONS RMQ & WEB API (`translationWebApi`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Summary | `summary` | 翻译数据服务 | [001_summary.md](../docs/02_translations/translations-rmq-web-api/001_summary.md) |
| 2 | Access Method | `access` | 翻译数据服务 | [002_access.md](../docs/02_translations/translations-rmq-web-api/002_access.md) |
| 3 | Integration Guide | `integrationGuide` | 翻译数据服务 | [003_integrationguide.md](../docs/02_translations/translations-rmq-web-api/003_integrationguide.md) |
| 4 | Available Languages | `availableLanguages` | 翻译数据服务 | [004_availablelanguages.md](../docs/02_translations/translations-rmq-web-api/004_availablelanguages.md) |
| 5 | Language Model | `languageModel` | 翻译数据服务 | [005_languagemodel.md](../docs/02_translations/translations-rmq-web-api/005_languagemodel.md) |
| 6 | Translation Response Model | `translationResponseModel` | 翻译数据服务 | [006_translationresponsemodel.md](../docs/02_translations/translations-rmq-web-api/006_translationresponsemodel.md) |
| 7 | Translation Model | `translationModel` | 翻译数据服务 | [007_translationmodel.md](../docs/02_translations/translations-rmq-web-api/007_translationmodel.md) |
| 8 | Message Syntax | `messageSyntax` | 翻译数据服务 | [008_messagesyntax.md](../docs/02_translations/translations-rmq-web-api/008_messagesyntax.md) |
| 9 | Web methods | `webMethods` | 翻译数据服务 | [009_webmethods.md](../docs/02_translations/translations-rmq-web-api/009_webmethods.md) |
| 10 | Samples | `samples` | 翻译数据服务 | [010_samples.md](../docs/02_translations/translations-rmq-web-api/010_samples.md) |
| 11 | Objects as part of the feed | `objectsAsPartFeed` | 翻译数据服务 | [011_objectsaspartfeed.md](../docs/02_translations/translations-rmq-web-api/011_objectsaspartfeed.md) |
| 12 | Change Log | `changeLog` | 翻译数据服务 | [012_changelog.md](../docs/02_translations/translations-rmq-web-api/012_changelog.md) |

### MATCH LIFECYCLE (`match_lifecycle_for_live`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Match Lifecycle | `root` | 体育数据模型参考 | [001_match-lifecycle.md](../docs/03_sports_model_reference/match-lifecycle/001_match-lifecycle.md) |

### EVENT TYPES (`eventTypes`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Periods | `periods` | 体育数据模型参考 | [001_periods.md](../docs/03_sports_model_reference/event-types/001_periods.md) |
| 2 | Event Types | `eventTypes` | 体育数据模型参考 | [002_eventtypes.md](../docs/03_sports_model_reference/event-types/002_eventtypes.md) |

### MARKET TYPES (`marketTypes`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Market Types | `root` | 体育数据模型参考 | [001_market-types.md](../docs/03_sports_model_reference/market-types/001_market-types.md) |

### SPORTS (`sports`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Sports | `root` | 体育数据模型参考 | [001_sports.md](../docs/03_sports_model_reference/sports/001_sports.md) |

### SPORTSBOOK NOTES (`sportsBookNotes`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | SportsBook Notes | `root` | 体育博彩业务规则 | [001_sportsbook-notes.md](../docs/04_sportsbook_rules/sportsbook-notes/001_sportsbook-notes.md) |

### SPORTS RULES (`sports_rules`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Settlement of Bets (All Sports) | `all-sports` | 体育博彩业务规则 | [001_all-sports.md](../docs/04_sportsbook_rules/sports-rules/001_all-sports.md) |
| 2 | Individual sports rules | `american-football` | 体育博彩业务规则 | [002_american-football.md](../docs/04_sportsbook_rules/sports-rules/002_american-football.md) |
| 3 | Badminton | `badminton` | 体育博彩业务规则 | [003_badminton.md](../docs/04_sportsbook_rules/sports-rules/003_badminton.md) |
| 4 | Bandy | `bandy` | 体育博彩业务规则 | [004_bandy.md](../docs/04_sportsbook_rules/sports-rules/004_bandy.md) |
| 5 | Baseball | `baseball` | 体育博彩业务规则 | [005_baseball.md](../docs/04_sportsbook_rules/sports-rules/005_baseball.md) |
| 6 | Basketball | `basketball` | 体育博彩业务规则 | [006_basketball.md](../docs/04_sportsbook_rules/sports-rules/006_basketball.md) |
| 7 | Beach Football/Soccer | `beach-football-soccer` | 体育博彩业务规则 | [007_beach-football-soccer.md](../docs/04_sportsbook_rules/sports-rules/007_beach-football-soccer.md) |
| 8 | Beach Volleyball | `beach-volleyball` | 体育博彩业务规则 | [008_beach-volleyball.md](../docs/04_sportsbook_rules/sports-rules/008_beach-volleyball.md) |
| 9 | Bowls | `bowls` | 体育博彩业务规则 | [009_bowls.md](../docs/04_sportsbook_rules/sports-rules/009_bowls.md) |
| 10 | Boxing/MMA/UFC | `boxing-mma-ufc` | 体育博彩业务规则 | [010_boxing-mma-ufc.md](../docs/04_sportsbook_rules/sports-rules/010_boxing-mma-ufc.md) |
| 11 | Boxing | `boxing` | 体育博彩业务规则 | [011_boxing.md](../docs/04_sportsbook_rules/sports-rules/011_boxing.md) |
| 12 | MMA | `mma` | 体育博彩业务规则 | [012_mma.md](../docs/04_sportsbook_rules/sports-rules/012_mma.md) |
| 13 | Cricket | `cricket` | 体育博彩业务规则 | [013_cricket.md](../docs/04_sportsbook_rules/sports-rules/013_cricket.md) |
| 14 | Cycling | `cycling` | 体育博彩业务规则 | [014_cycling.md](../docs/04_sportsbook_rules/sports-rules/014_cycling.md) |
| 15 | Darts | `darts` | 体育博彩业务规则 | [015_darts.md](../docs/04_sportsbook_rules/sports-rules/015_darts.md) |
| 16 | E-Sports | `e-sports` | 体育博彩业务规则 | [016_e-sports.md](../docs/04_sportsbook_rules/sports-rules/016_e-sports.md) |
| 17 | StarCraft II | `starcraft-2` | 体育博彩业务规则 | [017_starcraft-2.md](../docs/04_sportsbook_rules/sports-rules/017_starcraft-2.md) |
| 18 | Counter Strike 2 | `counter-strike-2` | 体育博彩业务规则 | [018_counter-strike-2.md](../docs/04_sportsbook_rules/sports-rules/018_counter-strike-2.md) |
| 19 | League of Legends (LOL) | `lague-of-legends` | 体育博彩业务规则 | [019_lague-of-legends.md](../docs/04_sportsbook_rules/sports-rules/019_lague-of-legends.md) |
| 20 | Dota 2 | `dota-2` | 体育博彩业务规则 | [020_dota-2.md](../docs/04_sportsbook_rules/sports-rules/020_dota-2.md) |
| 21 | HearthStone | `Hearth-stone` | 体育博彩业务规则 | [021_hearth-stone.md](../docs/04_sportsbook_rules/sports-rules/021_hearth-stone.md) |
| 22 | Call of Duty | `Call of Duty` | 体育博彩业务规则 | [022_call-of-duty.md](../docs/04_sportsbook_rules/sports-rules/022_call-of-duty.md) |
| 23 | Rocket League | `rocket-league ` | 体育博彩业务规则 | [023_rocket-league.md](../docs/04_sportsbook_rules/sports-rules/023_rocket-league.md) |
| 24 | King of Glory | `king-of-glory` | 体育博彩业务规则 | [024_king-of-glory.md](../docs/04_sportsbook_rules/sports-rules/024_king-of-glory.md) |
| 25 | Overwatch 2 | `overwatch-2` | 体育博彩业务规则 | [025_overwatch-2.md](../docs/04_sportsbook_rules/sports-rules/025_overwatch-2.md) |
| 26 | Rainbow 6 | `rainbow-6` | 体育博彩业务规则 | [026_rainbow-6.md](../docs/04_sportsbook_rules/sports-rules/026_rainbow-6.md) |
| 27 | Valorant | `valorant` | 体育博彩业务规则 | [027_valorant.md](../docs/04_sportsbook_rules/sports-rules/027_valorant.md) |
| 28 | World of Warcraft | `world-of-warcraft` | 体育博彩业务规则 | [028_world-of-warcraft.md](../docs/04_sportsbook_rules/sports-rules/028_world-of-warcraft.md) |
| 29 | Age of Empires | `age-of-empires` | 体育博彩业务规则 | [029_age-of-empires.md](../docs/04_sportsbook_rules/sports-rules/029_age-of-empires.md) |
| 30 | Apex Legends | `Apex Legends` | 体育博彩业务规则 | [030_apex-legends.md](../docs/04_sportsbook_rules/sports-rules/030_apex-legends.md) |
| 31 | Arena of Valor | `arena-of-valor` | 体育博彩业务规则 | [031_arena-of-valor.md](../docs/04_sportsbook_rules/sports-rules/031_arena-of-valor.md) |
| 32 | Artifact | `Artifact` | 体育博彩业务规则 | [032_artifact.md](../docs/04_sportsbook_rules/sports-rules/032_artifact.md) |
| 33 | Brawl Stars | `brawl-stars` | 体育博彩业务规则 | [033_brawl-stars.md](../docs/04_sportsbook_rules/sports-rules/033_brawl-stars.md) |
| 34 | PlayerUnknown's Battlegrounds (PUBG) | `pubg` | 体育博彩业务规则 | [034_pubg.md](../docs/04_sportsbook_rules/sports-rules/034_pubg.md) |
| 35 | PlayerUnknown's Battlegrounds (PUBG ) Mobile | `pubg-mobile` | 体育博彩业务规则 | [035_pubg-mobile.md](../docs/04_sportsbook_rules/sports-rules/035_pubg-mobile.md) |
| 36 | Deadlock | `deadlock` | 体育博彩业务规则 | [036_deadlock.md](../docs/04_sportsbook_rules/sports-rules/036_deadlock.md) |
| 37 | Clash of Clans | `clash-of-clans` | 体育博彩业务规则 | [037_clash-of-clans.md](../docs/04_sportsbook_rules/sports-rules/037_clash-of-clans.md) |
| 38 | Clash Royale | `clash-royale` | 体育博彩业务规则 | [038_clash-royale.md](../docs/04_sportsbook_rules/sports-rules/038_clash-royale.md) |
| 39 | Cross Fire | `cross-fire` | 体育博彩业务规则 | [039_cross-fire.md](../docs/04_sportsbook_rules/sports-rules/039_cross-fire.md) |
| 40 | Cross Fire HD league | `cross-fire-hd-league` | 体育博彩业务规则 | [040_cross-fire-hd-league.md](../docs/04_sportsbook_rules/sports-rules/040_cross-fire-hd-league.md) |
| 41 | CrossFire Mobile | `crossFire-mobile` | 体育博彩业务规则 | [041_crossfire-mobile.md](../docs/04_sportsbook_rules/sports-rules/041_crossfire-mobile.md) |
| 42 | Fortnite | `fortnite` | 体育博彩业务规则 | [042_fortnite.md](../docs/04_sportsbook_rules/sports-rules/042_fortnite.md) |
| 43 | Free Fire | `free-fire` | 体育博彩业务规则 | [043_free-fire.md](../docs/04_sportsbook_rules/sports-rules/043_free-fire.md) |
| 44 | Gwent | `gwent` | 体育博彩业务规则 | [044_gwent.md](../docs/04_sportsbook_rules/sports-rules/044_gwent.md) |
| 45 | Gears of War | `gears-of-war` | 体育博彩业务规则 | [045_gears-of-war.md](../docs/04_sportsbook_rules/sports-rules/045_gears-of-war.md) |
| 46 | Halo | `halo` | 体育博彩业务规则 | [046_halo.md](../docs/04_sportsbook_rules/sports-rules/046_halo.md) |
| 47 | Heroes of Newerth | `Heroes-of-newerth` | 体育博彩业务规则 | [047_heroes-of-newerth.md](../docs/04_sportsbook_rules/sports-rules/047_heroes-of-newerth.md) |
| 48 | League of Legends: Wild Rift | `league-of-legends-wild-rift` | 体育博彩业务规则 | [048_league-of-legends-wild-rift.md](../docs/04_sportsbook_rules/sports-rules/048_league-of-legends-wild-rift.md) |
| 49 | World of Tanks | `world-of-tanks` | 体育博彩业务规则 | [049_world-of-tanks.md](../docs/04_sportsbook_rules/sports-rules/049_world-of-tanks.md) |
| 50 | Mobile Legends | `mobile-legends` | 体育博彩业务规则 | [050_mobile-legends.md](../docs/04_sportsbook_rules/sports-rules/050_mobile-legends.md) |
| 51 | Floorball | `floorball` | 体育博彩业务规则 | [051_floorball.md](../docs/04_sportsbook_rules/sports-rules/051_floorball.md) |
| 52 | Football/Soccer | `football-soccer` | 体育博彩业务规则 | [052_football-soccer.md](../docs/04_sportsbook_rules/sports-rules/052_football-soccer.md) |
| 53 | Mixed/Mythical Football | `mixed-mythical-football` | 体育博彩业务规则 | [053_mixed-mythical-football.md](../docs/04_sportsbook_rules/sports-rules/053_mixed-mythical-football.md) |
| 54 | Futsal | `futsal` | 体育博彩业务规则 | [054_futsal.md](../docs/04_sportsbook_rules/sports-rules/054_futsal.md) |
| 55 | Irish/GAA Sports (Gaelic Football/Hurling) | `irish-gga-sports` | 体育博彩业务规则 | [055_irish-gga-sports.md](../docs/04_sportsbook_rules/sports-rules/055_irish-gga-sports.md) |
| 56 | Golf | `golf` | 体育博彩业务规则 | [056_golf.md](../docs/04_sportsbook_rules/sports-rules/056_golf.md) |
| 57 | Greyhound Racing | `greyhound-racing ` | 体育博彩业务规则 | [057_greyhound-racing.md](../docs/04_sportsbook_rules/sports-rules/057_greyhound-racing.md) |
| 58 | Handball | `handball` | 体育博彩业务规则 | [058_handball.md](../docs/04_sportsbook_rules/sports-rules/058_handball.md) |
| 59 | Hockey (Non-Ice, including ’Field’, ‘Rink’ or ‘Inline’ Hockey). | `hockey-non-ice` | 体育博彩业务规则 | [059_hockey-non-ice.md](../docs/04_sportsbook_rules/sports-rules/059_hockey-non-ice.md) |
| 60 | Horse Racing | `horse-racing` | 体育博彩业务规则 | [060_horse-racing.md](../docs/04_sportsbook_rules/sports-rules/060_horse-racing.md) |
| 61 | Ice Hockey | `ice-hockey` | 体育博彩业务规则 | [061_ice-hockey.md](../docs/04_sportsbook_rules/sports-rules/061_ice-hockey.md) |
| 62 | Motor Racing (Cars) | `motor-racing-cars` | 体育博彩业务规则 | [062_motor-racing-cars.md](../docs/04_sportsbook_rules/sports-rules/062_motor-racing-cars.md) |
| 63 | Nascar/Busch Racing | `nascar-busch-racing` | 体育博彩业务规则 | [063_nascar-busch-racing.md](../docs/04_sportsbook_rules/sports-rules/063_nascar-busch-racing.md) |
| 64 | Rally | `rally` | 体育博彩业务规则 | [064_rally.md](../docs/04_sportsbook_rules/sports-rules/064_rally.md) |
| 65 | Motorbikes | `motorbikes` | 体育博彩业务规则 | [065_motorbikes.md](../docs/04_sportsbook_rules/sports-rules/065_motorbikes.md) |
| 66 | Netball | `netball` | 体育博彩业务规则 | [066_netball.md](../docs/04_sportsbook_rules/sports-rules/066_netball.md) |
| 67 | Olympics | `olympics` | 体育博彩业务规则 | [067_olympics.md](../docs/04_sportsbook_rules/sports-rules/067_olympics.md) |
| 68 | Padel | `padel` | 体育博彩业务规则 | [068_padel.md](../docs/04_sportsbook_rules/sports-rules/068_padel.md) |
| 69 | Poker | `poker` | 体育博彩业务规则 | [069_poker.md](../docs/04_sportsbook_rules/sports-rules/069_poker.md) |
| 70 | Pool | `pool` | 体育博彩业务规则 | [070_pool.md](../docs/04_sportsbook_rules/sports-rules/070_pool.md) |
| 71 | Rugby League | `rugby_league` | 体育博彩业务规则 | [071_rugby-league.md](../docs/04_sportsbook_rules/sports-rules/071_rugby-league.md) |
| 72 | Rugby Union | `rugby_union` | 体育博彩业务规则 | [072_rugby-union.md](../docs/04_sportsbook_rules/sports-rules/072_rugby-union.md) |
| 73 | Snooker | `snooker` | 体育博彩业务规则 | [073_snooker.md](../docs/04_sportsbook_rules/sports-rules/073_snooker.md) |
| 74 | Speedway | `speedway` | 体育博彩业务规则 | [074_speedway.md](../docs/04_sportsbook_rules/sports-rules/074_speedway.md) |
| 75 | Squash | `squash` | 体育博彩业务规则 | [075_squash.md](../docs/04_sportsbook_rules/sports-rules/075_squash.md) |
| 76 | Table Tennis | `table-tennis` | 体育博彩业务规则 | [076_table-tennis.md](../docs/04_sportsbook_rules/sports-rules/076_table-tennis.md) |
| 77 | Tennis | `tennis` | 体育博彩业务规则 | [077_tennis.md](../docs/04_sportsbook_rules/sports-rules/077_tennis.md) |
| 78 | Volleyball | `volleyball` | 体育博彩业务规则 | [078_volleyball.md](../docs/04_sportsbook_rules/sports-rules/078_volleyball.md) |
| 79 | Water Polo | `water-polo` | 体育博彩业务规则 | [079_water-polo.md](../docs/04_sportsbook_rules/sports-rules/079_water-polo.md) |
| 80 | Winter Sports | `winter-sports` | 体育博彩业务规则 | [080_winter-sports.md](../docs/04_sportsbook_rules/sports-rules/080_winter-sports.md) |
| 81 | Other Sports | `other-sports` | 体育博彩业务规则 | [081_other-sports.md](../docs/04_sportsbook_rules/sports-rules/081_other-sports.md) |

### ODDS CONVERSION (`odds_conversion`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Odds Conversion | `root` | 赔率与计算 | [001_odds-conversion.md](../docs/05_odds_math/odds-conversion/001_odds-conversion.md) |

### BETGUARD (`betGuard`)

| 序号 | 标题 | location | 业务域 | Markdown |
|---:|---|---|---|---|
| 1 | Introduction | `introduction` | 投注风控服务 / BetGuard | [001_introduction.md](../docs/06_betguard_risk/betguard/001_introduction.md) |
| 2 | Bet Placement process | `bet_placement_process` | 投注风控服务 / BetGuard | [002_bet-placement-process.md](../docs/06_betguard_risk/betguard/002_bet-placement-process.md) |
| 3 | Bet Placement Flowchart | `bet_placement_flowchart` | 投注风控服务 / BetGuard | [003_bet-placement-flowchart.md](../docs/06_betguard_risk/betguard/003_bet-placement-flowchart.md) |
| 4 | Resulting Bets & Reporting Selection Outcomes | `resulting_bets_reporting_selection_outcomes` | 投注风控服务 / BetGuard | [004_resulting-bets-reporting-selection-outcomes.md](../docs/06_betguard_risk/betguard/004_resulting-bets-reporting-selection-outcomes.md) |
| 5 | BetGuard Notes | `betGuard_notes` | 投注风控服务 / BetGuard | [005_betguard-notes.md](../docs/06_betguard_risk/betguard/005_betguard-notes.md) |
| 6 | AuthToken | `authToken` | 投注风控服务 / BetGuard | [006_authtoken.md](../docs/06_betguard_risk/betguard/006_authtoken.md) |
| 7 | Hash Parameter | `hash_parameter` | 投注风控服务 / BetGuard | [007_hash-parameter.md](../docs/06_betguard_risk/betguard/007_hash-parameter.md) |
| 8 | Hash Calculation | `hash_calculation` | 投注风控服务 / BetGuard | [008_hash-calculation.md](../docs/06_betguard_risk/betguard/008_hash-calculation.md) |
| 9 | Example | `example` | 投注风控服务 / BetGuard | [009_example.md](../docs/06_betguard_risk/betguard/009_example.md) |
| 10 | Partner API Security Checks | `partner_api_security_checks` | 投注风控服务 / BetGuard | [010_partner-api-security-checks.md](../docs/06_betguard_risk/betguard/010_partner-api-security-checks.md) |
| 11 | TS Parameter (Timestamp) | `ts_parameter_timestamp` | 投注风控服务 / BetGuard | [011_ts-parameter-timestamp.md](../docs/06_betguard_risk/betguard/011_ts-parameter-timestamp.md) |
| 12 | Hash Parameter | `partner_api_hash_parameter` | 投注风控服务 / BetGuard | [012_partner-api-hash-parameter.md](../docs/06_betguard_risk/betguard/012_partner-api-hash-parameter.md) |
| 13 | Hash Calculation | `partner_api_hash_calculation` | 投注风控服务 / BetGuard | [013_partner-api-hash-calculation.md](../docs/06_betguard_risk/betguard/013_partner-api-hash-calculation.md) |
| 14 | Hash Parameters Lists per Method | `hash_parameters_lists_per_method` | 投注风控服务 / BetGuard | [014_hash-parameters-lists-per-method.md](../docs/06_betguard_risk/betguard/014_hash-parameters-lists-per-method.md) |
| 15 | Example | `partner_api_example` | 投注风控服务 / BetGuard | [015_partner-api-example.md](../docs/06_betguard_risk/betguard/015_partner-api-example.md) |
| 16 | AuthToken; Binding with Currency | `authToken_binding_with_currency` | 投注风控服务 / BetGuard | [016_authtoken-binding-with-currency.md](../docs/06_betguard_risk/betguard/016_authtoken-binding-with-currency.md) |
| 17 | BetGuard API Calls | `betGuard_api` | 投注风控服务 / BetGuard | [017_betguard-api.md](../docs/06_betguard_risk/betguard/017_betguard-api.md) |
| 18 | CreateBet | `createBet` | 投注风控服务 / BetGuard | [018_createbet.md](../docs/06_betguard_risk/betguard/018_createbet.md) |
| 19 | CreateBet Request Sample | `createBet_request_sample` | 投注风控服务 / BetGuard | [019_createbet-request-sample.md](../docs/06_betguard_risk/betguard/019_createbet-request-sample.md) |
| 20 | CreateBet Response Sample | `createBet_response_sample` | 投注风控服务 / BetGuard | [020_createbet-response-sample.md](../docs/06_betguard_risk/betguard/020_createbet-response-sample.md) |
| 21 | GetMaxBetAmount | `get_max_bet_amount` | 投注风控服务 / BetGuard | [021_get-max-bet-amount.md](../docs/06_betguard_risk/betguard/021_get-max-bet-amount.md) |
| 22 | GetMaxBetAmount Request Sample | `get_max_bet_amount_request_sample` | 投注风控服务 / BetGuard | [022_get-max-bet-amount-request-sample.md](../docs/06_betguard_risk/betguard/022_get-max-bet-amount-request-sample.md) |
| 23 | GetMaxBetAmount Response Sample | `get_max_bet_amount_response_sample` | 投注风控服务 / BetGuard | [023_get-max-bet-amount-response-sample.md](../docs/06_betguard_risk/betguard/023_get-max-bet-amount-response-sample.md) |
| 24 | ResendFailedTransfers | `resend_failed_transfers` | 投注风控服务 / BetGuard | [024_resend-failed-transfers.md](../docs/06_betguard_risk/betguard/024_resend-failed-transfers.md) |
| 25 | ResendFailedTransfers Request Sample | `resend_failed_transfers_request_sample` | 投注风控服务 / BetGuard | [025_resend-failed-transfers-request-sample.md](../docs/06_betguard_risk/betguard/025_resend-failed-transfers-request-sample.md) |
| 26 | ResendFailedTransfers Response Sample | `resend_failed_transfers_response_sample` | 投注风控服务 / BetGuard | [026_resend-failed-transfers-response-sample.md](../docs/06_betguard_risk/betguard/026_resend-failed-transfers-response-sample.md) |
| 27 | MarkBetAsCashout | `mark_bet_as_cashout` | 投注风控服务 / BetGuard | [027_mark-bet-as-cashout.md](../docs/06_betguard_risk/betguard/027_mark-bet-as-cashout.md) |
| 28 | MarkBetAsCashout Request Sample | `mark_bet_as_cashout_request_sample` | 投注风控服务 / BetGuard | [028_mark-bet-as-cashout-request-sample.md](../docs/06_betguard_risk/betguard/028_mark-bet-as-cashout-request-sample.md) |
| 29 | MarkBetAsCashout Response Sample | `mark_bet_as_cashout_response_sample` | 投注风控服务 / BetGuard | [029_mark-bet-as-cashout-response-sample.md](../docs/06_betguard_risk/betguard/029_mark-bet-as-cashout-response-sample.md) |
| 30 | CheckAndMarkBetAsCashout | `check_and_mark_bet_as_cashout` | 投注风控服务 / BetGuard | [030_check-and-mark-bet-as-cashout.md](../docs/06_betguard_risk/betguard/030_check-and-mark-bet-as-cashout.md) |
| 31 | CheckAndMarkBetAsCashout Request Sample | `check_and_mark_bet_as_cashout_request_sample` | 投注风控服务 / BetGuard | [031_check-and-mark-bet-as-cashout-request-sample.md](../docs/06_betguard_risk/betguard/031_check-and-mark-bet-as-cashout-request-sample.md) |
| 32 | CheckAndMarkBetAsCashout Response Sample | `check_and_mark_bet_as_cashout_response_sample` | 投注风控服务 / BetGuard | [032_check-and-mark-bet-as-cashout-response-sample.md](../docs/06_betguard_risk/betguard/032_check-and-mark-bet-as-cashout-response-sample.md) |
| 33 | ReturnBet | `return_bet` | 投注风控服务 / BetGuard | [033_return-bet.md](../docs/06_betguard_risk/betguard/033_return-bet.md) |
| 34 | ReturnBet Request Sample | `return_bet_request_sample` | 投注风控服务 / BetGuard | [034_return-bet-request-sample.md](../docs/06_betguard_risk/betguard/034_return-bet-request-sample.md) |
| 35 | ReturnBet Response Sample | `return_bet_response_sample` | 投注风控服务 / BetGuard | [035_return-bet-response-sample.md](../docs/06_betguard_risk/betguard/035_return-bet-response-sample.md) |
| 36 | UpdateClient | `update_client` | 投注风控服务 / BetGuard | [036_update-client.md](../docs/06_betguard_risk/betguard/036_update-client.md) |
| 37 | UpdateClient Request Sample | `update_client_request_sample` | 投注风控服务 / BetGuard | [037_update-client-request-sample.md](../docs/06_betguard_risk/betguard/037_update-client-request-sample.md) |
| 38 | UpdateClient Response Sample | `update_client_response_sample` | 投注风控服务 / BetGuard | [038_update-client-response-sample.md](../docs/06_betguard_risk/betguard/038_update-client-response-sample.md) |
| 39 | GetClientDetails | `partner_api_get_client_details` | 投注风控服务 / BetGuard | [039_partner-api-get-client-details.md](../docs/06_betguard_risk/betguard/039_partner-api-get-client-details.md) |
| 40 | GetClientDetails Request Sample | `partner_api_request_sample` | 投注风控服务 / BetGuard | [040_partner-api-request-sample.md](../docs/06_betguard_risk/betguard/040_partner-api-request-sample.md) |
| 41 | GetClientDetails Response Sample | `partner_api_response_sample` | 投注风控服务 / BetGuard | [041_partner-api-response-sample.md](../docs/06_betguard_risk/betguard/041_partner-api-response-sample.md) |
| 42 | BetPlaced | `partner_api_bet_placed` | 投注风控服务 / BetGuard | [042_partner-api-bet-placed.md](../docs/06_betguard_risk/betguard/042_partner-api-bet-placed.md) |
| 43 | BetPlaced Request Sample | `bet_placed_request_sample` | 投注风控服务 / BetGuard | [043_bet-placed-request-sample.md](../docs/06_betguard_risk/betguard/043_bet-placed-request-sample.md) |
| 44 | BetPlaced Response Sample | `bet_placed_response_sample` | 投注风控服务 / BetGuard | [044_bet-placed-response-sample.md](../docs/06_betguard_risk/betguard/044_bet-placed-response-sample.md) |
| 45 | BetResulted | `bet_resulted` | 投注风控服务 / BetGuard | [045_bet-resulted.md](../docs/06_betguard_risk/betguard/045_bet-resulted.md) |
| 46 | BetResulted Request Sample | `bet_resulted_request_sample` | 投注风控服务 / BetGuard | [046_bet-resulted-request-sample.md](../docs/06_betguard_risk/betguard/046_bet-resulted-request-sample.md) |
| 47 | BetResulted Response Sample | `bet_resulted_response_sample` | 投注风控服务 / BetGuard | [047_bet-resulted-response-sample.md](../docs/06_betguard_risk/betguard/047_bet-resulted-response-sample.md) |
| 48 | BetResulted Retry Logic | `bet_resulted_retry_logic` | 投注风控服务 / BetGuard | [048_bet-resulted-retry-logic.md](../docs/06_betguard_risk/betguard/048_bet-resulted-retry-logic.md) |
| 49 | Rollback | `partner_api_rollback` | 投注风控服务 / BetGuard | [049_partner-api-rollback.md](../docs/06_betguard_risk/betguard/049_partner-api-rollback.md) |
| 50 | Rollback Request Sample | `rollback_request_sample` | 投注风控服务 / BetGuard | [050_rollback-request-sample.md](../docs/06_betguard_risk/betguard/050_rollback-request-sample.md) |
| 51 | Rollback Response Sample | `rollback_response_sample` | 投注风控服务 / BetGuard | [051_rollback-response-sample.md](../docs/06_betguard_risk/betguard/051_rollback-response-sample.md) |
| 52 | Rollback Retry Logic | `rollback_retry_logic` | 投注风控服务 / BetGuard | [052_rollback-retry-logic.md](../docs/06_betguard_risk/betguard/052_rollback-retry-logic.md) |
| 53 | Client | `partner_api_client` | 投注风控服务 / BetGuard | [053_partner-api-client.md](../docs/06_betguard_risk/betguard/053_partner-api-client.md) |
| 54 | Bet Selection | `partner_api_bet_selection` | 投注风控服务 / BetGuard | [054_partner-api-bet-selection.md](../docs/06_betguard_risk/betguard/054_partner-api-bet-selection.md) |
| 55 | Errors returned by the FeedConstruct | `errors_returned_by_the_feed_construct` | 投注风控服务 / BetGuard | [055_errors-returned-by-the-feed-construct.md](../docs/06_betguard_risk/betguard/055_errors-returned-by-the-feed-construct.md) |
| 56 | Errors returned by the Partner | `errors_returned_by_the_partner` | 投注风控服务 / BetGuard | [056_errors-returned-by-the-partner.md](../docs/06_betguard_risk/betguard/056_errors-returned-by-the-partner.md) |
| 57 | Bet Limits - Global Limits | `bet_limits_global_limits` | 投注风控服务 / BetGuard | [057_bet-limits-global-limits.md](../docs/06_betguard_risk/betguard/057_bet-limits-global-limits.md) |
| 58 | Client Default | `client_default` | 投注风控服务 / BetGuard | [058_client-default.md](../docs/06_betguard_risk/betguard/058_client-default.md) |
| 59 | Multiple Bets | `multiple_bets` | 投注风控服务 / BetGuard | [059_multiple-bets.md](../docs/06_betguard_risk/betguard/059_multiple-bets.md) |
| 60 | Reset Time and Global Live Delay | `reset_time_and_global_live_delay` | 投注风控服务 / BetGuard | [060_reset-time-and-global-live-delay.md](../docs/06_betguard_risk/betguard/060_reset-time-and-global-live-delay.md) |
| 61 | Bet Limits - Sport Limits | `bet_limits_sport_limits` | 投注风控服务 / BetGuard | [061_bet-limits-sport-limits.md](../docs/06_betguard_risk/betguard/061_bet-limits-sport-limits.md) |
| 62 | How to apply limits? | `how_to_apply_limits` | 投注风控服务 / BetGuard | [062_how-to-apply-limits.md](../docs/06_betguard_risk/betguard/062_how-to-apply-limits.md) |
| 63 | Currency Codes | `currency_codes` | 投注风控服务 / BetGuard | [063_currency-codes.md](../docs/06_betguard_risk/betguard/063_currency-codes.md) |
| 64 | .Net/C# | `cSharp` | 投注风控服务 / BetGuard | [064_csharp.md](../docs/06_betguard_risk/betguard/064_csharp.md) |
| 65 | Change Log | `betGuard_change_log` | 投注风控服务 / BetGuard | [065_betguard-change-log.md](../docs/06_betguard_risk/betguard/065_betguard-change-log.md) |
