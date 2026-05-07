# BC 文档业务域索引

该索引按体育博彩数据商常见能力分层：数据源服务、翻译数据、体育模型参考、Sportsbook 规则、赔率计算与 BetGuard 投注风控。

## 数据源服务 / OddsFeed

| 模块 | 标题 | location | Markdown |
|---|---|---|---|
| RABBIT MQ | Summary | `summary` | [001_summary.md](../docs/01_data_feed/rmq-web-api/001_summary.md) |
| RABBIT MQ | Access method | `access` | [002_access.md](../docs/01_data_feed/rmq-web-api/002_access.md) |
| RABBIT MQ | Integration Guide | `integrationGuide` | [003_integrationguide.md](../docs/01_data_feed/rmq-web-api/003_integrationguide.md) |
| RABBIT MQ | Objects Descriptions | `objectsDescriptions` | [004_objectsdescriptions.md](../docs/01_data_feed/rmq-web-api/004_objectsdescriptions.md) |
| RABBIT MQ | Sport | `sport` | [005_sport.md](../docs/01_data_feed/rmq-web-api/005_sport.md) |
| RABBIT MQ | Region | `region` | [006_region.md](../docs/01_data_feed/rmq-web-api/006_region.md) |
| RABBIT MQ | Competition | `competition` | [007_competition.md](../docs/01_data_feed/rmq-web-api/007_competition.md) |
| RABBIT MQ | Match | `match` | [008_match.md](../docs/01_data_feed/rmq-web-api/008_match.md) |
| RABBIT MQ | Market | `market` | [009_market.md](../docs/01_data_feed/rmq-web-api/009_market.md) |
| RABBIT MQ | Selection | `selection` | [010_selection.md](../docs/01_data_feed/rmq-web-api/010_selection.md) |
| RABBIT MQ | Stat | `stat` | [011_stat.md](../docs/01_data_feed/rmq-web-api/011_stat.md) |
| RABBIT MQ | MatchMember | `matchMember` | [012_matchmember.md](../docs/01_data_feed/rmq-web-api/012_matchmember.md) |
| RABBIT MQ | MarketType | `marketType` | [013_markettype.md](../docs/01_data_feed/rmq-web-api/013_markettype.md) |
| RABBIT MQ | SelectionType | `selectionType` | [014_selectiontype.md](../docs/01_data_feed/rmq-web-api/014_selectiontype.md) |
| RABBIT MQ | VoidNotification | `voidNotification` | [015_voidnotification.md](../docs/01_data_feed/rmq-web-api/015_voidnotification.md) |
| RABBIT MQ | PartnerBooking | `partnerBooking` | [016_partnerbooking.md](../docs/01_data_feed/rmq-web-api/016_partnerbooking.md) |
| RABBIT MQ | Calendar Match | `calendar-match` | [017_calendar-match.md](../docs/01_data_feed/rmq-web-api/017_calendar-match.md) |
| RABBIT MQ | Book | `book` | [018_book.md](../docs/01_data_feed/rmq-web-api/018_book.md) |
| RABBIT MQ | EventType | `event-types` | [019_event-types.md](../docs/01_data_feed/rmq-web-api/019_event-types.md) |
| RABBIT MQ | Period | `periods` | [020_periods.md](../docs/01_data_feed/rmq-web-api/020_periods.md) |
| RABBIT MQ | UnBookObject | `unBookObject` | [021_unbookobject.md](../docs/01_data_feed/rmq-web-api/021_unbookobject.md) |
| RABBIT MQ | BookObject | `bookObject` | [022_bookobject.md](../docs/01_data_feed/rmq-web-api/022_bookobject.md) |
| RABBIT MQ | MarketTypeBooking | `marketTypeBooking` | [023_markettypebooking.md](../docs/01_data_feed/rmq-web-api/023_markettypebooking.md) |
| RABBIT MQ | RacingInfo | `racingInfo` | [024_racinginfo.md](../docs/01_data_feed/rmq-web-api/024_racinginfo.md) |
| RABBIT MQ | MarketExtraInfo | `marketExtraInfo` | [025_marketextrainfo.md](../docs/01_data_feed/rmq-web-api/025_marketextrainfo.md) |
| RABBIT MQ | BetDivident | `betDivident` | [026_betdivident.md](../docs/01_data_feed/rmq-web-api/026_betdivident.md) |
| RABBIT MQ | ManualEachWayTerms | `manualEachWayTerms` | [027_manualeachwayterms.md](../docs/01_data_feed/rmq-web-api/027_manualeachwayterms.md) |
| RABBIT MQ | HorseInfo | `horseInfo` | [028_horseinfo.md](../docs/01_data_feed/rmq-web-api/028_horseinfo.md) |
| RABBIT MQ | SportOrder | `sportOrder` | [029_sportorder.md](../docs/01_data_feed/rmq-web-api/029_sportorder.md) |
| RABBIT MQ | CommandResponse | `commandResponse` | [030_commandresponse.md](../docs/01_data_feed/rmq-web-api/030_commandresponse.md) |
| RABBIT MQ | ResponseError | `responseError` | [031_responseerror.md](../docs/01_data_feed/rmq-web-api/031_responseerror.md) |
| RABBIT MQ | Message Syntax | `messageSyntax` | [032_messagesyntax.md](../docs/01_data_feed/rmq-web-api/032_messagesyntax.md) |
| RABBIT MQ | Web Methods | `webMethods` | [033_webmethods.md](../docs/01_data_feed/rmq-web-api/033_webmethods.md) |
| RABBIT MQ | Objects as part of the feed | `objectsAsPartFeed` | [034_objectsaspartfeed.md](../docs/01_data_feed/rmq-web-api/034_objectsaspartfeed.md) |
| RABBIT MQ | Object Types | `objectTypes` | [035_objecttypes.md](../docs/01_data_feed/rmq-web-api/035_objecttypes.md) |
| RABBIT MQ | Integration Notes | `integrationNotes` | [036_integrationnotes.md](../docs/01_data_feed/rmq-web-api/036_integrationnotes.md) |
| RABBIT MQ | Result Codes | `resultCodes` | [037_resultcodes.md](../docs/01_data_feed/rmq-web-api/037_resultcodes.md) |
| RABBIT MQ | Samples | `samples` | [038_samples.md](../docs/01_data_feed/rmq-web-api/038_samples.md) |
| RABBIT MQ | .Net/C# | `cSharp` | [039_csharp.md](../docs/01_data_feed/rmq-web-api/039_csharp.md) |
| RABBIT MQ | Node.js | `nodeJs` | [040_nodejs.md](../docs/01_data_feed/rmq-web-api/040_nodejs.md) |
| RABBIT MQ | PHP | `php` | [041_php.md](../docs/01_data_feed/rmq-web-api/041_php.md) |
| RABBIT MQ | Java | `java` | [042_java.md](../docs/01_data_feed/rmq-web-api/042_java.md) |
| RABBIT MQ | Change Log | `changeLog` | [043_changelog.md](../docs/01_data_feed/rmq-web-api/043_changelog.md) |
| TCP SOCKET | Summary | `summary` | [001_summary.md](../docs/01_data_feed/tcp-socket-api/001_summary.md) |
| TCP SOCKET | Access Method | `access` | [002_access.md](../docs/01_data_feed/tcp-socket-api/002_access.md) |
| TCP SOCKET | Integration Guide | `integrationGuide` | [003_integrationguide.md](../docs/01_data_feed/tcp-socket-api/003_integrationguide.md) |
| TCP SOCKET | Objects Descriptions | `objectsDescriptions` | [004_objectsdescriptions.md](../docs/01_data_feed/tcp-socket-api/004_objectsdescriptions.md) |
| TCP SOCKET | Sport | `sport` | [005_sport.md](../docs/01_data_feed/tcp-socket-api/005_sport.md) |
| TCP SOCKET | Region | `region` | [006_region.md](../docs/01_data_feed/tcp-socket-api/006_region.md) |
| TCP SOCKET | Competition | `competition` | [007_competition.md](../docs/01_data_feed/tcp-socket-api/007_competition.md) |
| TCP SOCKET | Match | `match` | [008_match.md](../docs/01_data_feed/tcp-socket-api/008_match.md) |
| TCP SOCKET | Market | `market` | [009_market.md](../docs/01_data_feed/tcp-socket-api/009_market.md) |
| TCP SOCKET | Selection | `selection` | [010_selection.md](../docs/01_data_feed/tcp-socket-api/010_selection.md) |
| TCP SOCKET | Stat | `stat` | [011_stat.md](../docs/01_data_feed/tcp-socket-api/011_stat.md) |
| TCP SOCKET | MatchMember | `matchMember` | [012_matchmember.md](../docs/01_data_feed/tcp-socket-api/012_matchmember.md) |
| TCP SOCKET | MarketType | `marketType` | [013_markettype.md](../docs/01_data_feed/tcp-socket-api/013_markettype.md) |
| TCP SOCKET | SelectionType | `selectionType` | [014_selectiontype.md](../docs/01_data_feed/tcp-socket-api/014_selectiontype.md) |
| TCP SOCKET | VoidNotification | `voidNotification` | [015_voidnotification.md](../docs/01_data_feed/tcp-socket-api/015_voidnotification.md) |
| TCP SOCKET | PartnerBooking | `partnerBooking` | [016_partnerbooking.md](../docs/01_data_feed/tcp-socket-api/016_partnerbooking.md) |
| TCP SOCKET | Calendar Match | `calendar-match` | [017_calendar-match.md](../docs/01_data_feed/tcp-socket-api/017_calendar-match.md) |
| TCP SOCKET | BookItems | `book-items` | [018_book-items.md](../docs/01_data_feed/tcp-socket-api/018_book-items.md) |
| TCP SOCKET | EventType | `event-types` | [019_event-types.md](../docs/01_data_feed/tcp-socket-api/019_event-types.md) |
| TCP SOCKET | Period | `periods` | [020_periods.md](../docs/01_data_feed/tcp-socket-api/020_periods.md) |
| TCP SOCKET | UnBookObject | `unBookObject` | [021_unbookobject.md](../docs/01_data_feed/tcp-socket-api/021_unbookobject.md) |
| TCP SOCKET | BookObject | `bookObject` | [022_bookobject.md](../docs/01_data_feed/tcp-socket-api/022_bookobject.md) |
| TCP SOCKET | MarketTypeBooking | `marketTypeBooking` | [023_markettypebooking.md](../docs/01_data_feed/tcp-socket-api/023_markettypebooking.md) |
| TCP SOCKET | SportOrder | `sportOrder` | [024_sportorder.md](../docs/01_data_feed/tcp-socket-api/024_sportorder.md) |
| TCP SOCKET | ResponseError | `responseError` | [025_responseerror.md](../docs/01_data_feed/tcp-socket-api/025_responseerror.md) |
| TCP SOCKET | Message Syntax | `messageSyntax` | [026_messagesyntax.md](../docs/01_data_feed/tcp-socket-api/026_messagesyntax.md) |
| TCP SOCKET | Command List | `commandList` | [027_commandlist.md](../docs/01_data_feed/tcp-socket-api/027_commandlist.md) |
| TCP SOCKET | Objects as part of the feed | `objectsAsPartFeed` | [028_objectsaspartfeed.md](../docs/01_data_feed/tcp-socket-api/028_objectsaspartfeed.md) |
| TCP SOCKET | Object Types | `objectTypes` | [029_objecttypes.md](../docs/01_data_feed/tcp-socket-api/029_objecttypes.md) |
| TCP SOCKET | Integration Notes | `integrationNotes` | [030_integrationnotes.md](../docs/01_data_feed/tcp-socket-api/030_integrationnotes.md) |
| TCP SOCKET | Result Codes | `resultCodes` | [031_resultcodes.md](../docs/01_data_feed/tcp-socket-api/031_resultcodes.md) |
| TCP SOCKET | Samples | `samples` | [032_samples.md](../docs/01_data_feed/tcp-socket-api/032_samples.md) |
| TCP SOCKET | Change Log | `changeLog` | [033_changelog.md](../docs/01_data_feed/tcp-socket-api/033_changelog.md) |

## 翻译数据服务

| 模块 | 标题 | location | Markdown |
|---|---|---|---|
| TRANSLATIONS SOCKET API | Summary | `summary` | [001_summary.md](../docs/02_translations/translations-socket-api/001_summary.md) |
| TRANSLATIONS SOCKET API | Access Method | `access` | [002_access.md](../docs/02_translations/translations-socket-api/002_access.md) |
| TRANSLATIONS SOCKET API | Integration Guide | `integrationGuide` | [003_integrationguide.md](../docs/02_translations/translations-socket-api/003_integrationguide.md) |
| TRANSLATIONS SOCKET API | Objects Descriptions | `objectsDescriptions` | [004_objectsdescriptions.md](../docs/02_translations/translations-socket-api/004_objectsdescriptions.md) |
| TRANSLATIONS SOCKET API | Message Syntax | `messageSyntax` | [005_messagesyntax.md](../docs/02_translations/translations-socket-api/005_messagesyntax.md) |
| TRANSLATIONS SOCKET API | Command List | `commandList` | [006_commandlist.md](../docs/02_translations/translations-socket-api/006_commandlist.md) |
| TRANSLATIONS SOCKET API | Objects as part of the feed | `objectsAsPartFeed` | [007_objectsaspartfeed.md](../docs/02_translations/translations-socket-api/007_objectsaspartfeed.md) |
| TRANSLATIONS SOCKET API | Change Log | `changeLog` | [008_changelog.md](../docs/02_translations/translations-socket-api/008_changelog.md) |
| TRANSLATIONS RMQ & WEB API | Summary | `summary` | [001_summary.md](../docs/02_translations/translations-rmq-web-api/001_summary.md) |
| TRANSLATIONS RMQ & WEB API | Access Method | `access` | [002_access.md](../docs/02_translations/translations-rmq-web-api/002_access.md) |
| TRANSLATIONS RMQ & WEB API | Integration Guide | `integrationGuide` | [003_integrationguide.md](../docs/02_translations/translations-rmq-web-api/003_integrationguide.md) |
| TRANSLATIONS RMQ & WEB API | Available Languages | `availableLanguages` | [004_availablelanguages.md](../docs/02_translations/translations-rmq-web-api/004_availablelanguages.md) |
| TRANSLATIONS RMQ & WEB API | Language Model | `languageModel` | [005_languagemodel.md](../docs/02_translations/translations-rmq-web-api/005_languagemodel.md) |
| TRANSLATIONS RMQ & WEB API | Translation Response Model | `translationResponseModel` | [006_translationresponsemodel.md](../docs/02_translations/translations-rmq-web-api/006_translationresponsemodel.md) |
| TRANSLATIONS RMQ & WEB API | Translation Model | `translationModel` | [007_translationmodel.md](../docs/02_translations/translations-rmq-web-api/007_translationmodel.md) |
| TRANSLATIONS RMQ & WEB API | Message Syntax | `messageSyntax` | [008_messagesyntax.md](../docs/02_translations/translations-rmq-web-api/008_messagesyntax.md) |
| TRANSLATIONS RMQ & WEB API | Web methods | `webMethods` | [009_webmethods.md](../docs/02_translations/translations-rmq-web-api/009_webmethods.md) |
| TRANSLATIONS RMQ & WEB API | Samples | `samples` | [010_samples.md](../docs/02_translations/translations-rmq-web-api/010_samples.md) |
| TRANSLATIONS RMQ & WEB API | Objects as part of the feed | `objectsAsPartFeed` | [011_objectsaspartfeed.md](../docs/02_translations/translations-rmq-web-api/011_objectsaspartfeed.md) |
| TRANSLATIONS RMQ & WEB API | Change Log | `changeLog` | [012_changelog.md](../docs/02_translations/translations-rmq-web-api/012_changelog.md) |

## 体育数据模型参考

| 模块 | 标题 | location | Markdown |
|---|---|---|---|
| MATCH LIFECYCLE | Match Lifecycle | `root` | [001_match-lifecycle.md](../docs/03_sports_model_reference/match-lifecycle/001_match-lifecycle.md) |
| EVENT TYPES | Periods | `periods` | [001_periods.md](../docs/03_sports_model_reference/event-types/001_periods.md) |
| EVENT TYPES | Event Types | `eventTypes` | [002_eventtypes.md](../docs/03_sports_model_reference/event-types/002_eventtypes.md) |
| MARKET TYPES | Market Types | `root` | [001_market-types.md](../docs/03_sports_model_reference/market-types/001_market-types.md) |
| SPORTS | Sports | `root` | [001_sports.md](../docs/03_sports_model_reference/sports/001_sports.md) |

## 体育博彩业务规则

| 模块 | 标题 | location | Markdown |
|---|---|---|---|
| SPORTSBOOK NOTES | SportsBook Notes | `root` | [001_sportsbook-notes.md](../docs/04_sportsbook_rules/sportsbook-notes/001_sportsbook-notes.md) |
| SPORTS RULES | Settlement of Bets (All Sports) | `all-sports` | [001_all-sports.md](../docs/04_sportsbook_rules/sports-rules/001_all-sports.md) |
| SPORTS RULES | Individual sports rules | `american-football` | [002_american-football.md](../docs/04_sportsbook_rules/sports-rules/002_american-football.md) |
| SPORTS RULES | Badminton | `badminton` | [003_badminton.md](../docs/04_sportsbook_rules/sports-rules/003_badminton.md) |
| SPORTS RULES | Bandy | `bandy` | [004_bandy.md](../docs/04_sportsbook_rules/sports-rules/004_bandy.md) |
| SPORTS RULES | Baseball | `baseball` | [005_baseball.md](../docs/04_sportsbook_rules/sports-rules/005_baseball.md) |
| SPORTS RULES | Basketball | `basketball` | [006_basketball.md](../docs/04_sportsbook_rules/sports-rules/006_basketball.md) |
| SPORTS RULES | Beach Football/Soccer | `beach-football-soccer` | [007_beach-football-soccer.md](../docs/04_sportsbook_rules/sports-rules/007_beach-football-soccer.md) |
| SPORTS RULES | Beach Volleyball | `beach-volleyball` | [008_beach-volleyball.md](../docs/04_sportsbook_rules/sports-rules/008_beach-volleyball.md) |
| SPORTS RULES | Bowls | `bowls` | [009_bowls.md](../docs/04_sportsbook_rules/sports-rules/009_bowls.md) |
| SPORTS RULES | Boxing/MMA/UFC | `boxing-mma-ufc` | [010_boxing-mma-ufc.md](../docs/04_sportsbook_rules/sports-rules/010_boxing-mma-ufc.md) |
| SPORTS RULES | Boxing | `boxing` | [011_boxing.md](../docs/04_sportsbook_rules/sports-rules/011_boxing.md) |
| SPORTS RULES | MMA | `mma` | [012_mma.md](../docs/04_sportsbook_rules/sports-rules/012_mma.md) |
| SPORTS RULES | Cricket | `cricket` | [013_cricket.md](../docs/04_sportsbook_rules/sports-rules/013_cricket.md) |
| SPORTS RULES | Cycling | `cycling` | [014_cycling.md](../docs/04_sportsbook_rules/sports-rules/014_cycling.md) |
| SPORTS RULES | Darts | `darts` | [015_darts.md](../docs/04_sportsbook_rules/sports-rules/015_darts.md) |
| SPORTS RULES | E-Sports | `e-sports` | [016_e-sports.md](../docs/04_sportsbook_rules/sports-rules/016_e-sports.md) |
| SPORTS RULES | StarCraft II | `starcraft-2` | [017_starcraft-2.md](../docs/04_sportsbook_rules/sports-rules/017_starcraft-2.md) |
| SPORTS RULES | Counter Strike 2 | `counter-strike-2` | [018_counter-strike-2.md](../docs/04_sportsbook_rules/sports-rules/018_counter-strike-2.md) |
| SPORTS RULES | League of Legends (LOL) | `lague-of-legends` | [019_lague-of-legends.md](../docs/04_sportsbook_rules/sports-rules/019_lague-of-legends.md) |
| SPORTS RULES | Dota 2 | `dota-2` | [020_dota-2.md](../docs/04_sportsbook_rules/sports-rules/020_dota-2.md) |
| SPORTS RULES | HearthStone | `Hearth-stone` | [021_hearth-stone.md](../docs/04_sportsbook_rules/sports-rules/021_hearth-stone.md) |
| SPORTS RULES | Call of Duty | `Call of Duty` | [022_call-of-duty.md](../docs/04_sportsbook_rules/sports-rules/022_call-of-duty.md) |
| SPORTS RULES | Rocket League | `rocket-league ` | [023_rocket-league.md](../docs/04_sportsbook_rules/sports-rules/023_rocket-league.md) |
| SPORTS RULES | King of Glory | `king-of-glory` | [024_king-of-glory.md](../docs/04_sportsbook_rules/sports-rules/024_king-of-glory.md) |
| SPORTS RULES | Overwatch 2 | `overwatch-2` | [025_overwatch-2.md](../docs/04_sportsbook_rules/sports-rules/025_overwatch-2.md) |
| SPORTS RULES | Rainbow 6 | `rainbow-6` | [026_rainbow-6.md](../docs/04_sportsbook_rules/sports-rules/026_rainbow-6.md) |
| SPORTS RULES | Valorant | `valorant` | [027_valorant.md](../docs/04_sportsbook_rules/sports-rules/027_valorant.md) |
| SPORTS RULES | World of Warcraft | `world-of-warcraft` | [028_world-of-warcraft.md](../docs/04_sportsbook_rules/sports-rules/028_world-of-warcraft.md) |
| SPORTS RULES | Age of Empires | `age-of-empires` | [029_age-of-empires.md](../docs/04_sportsbook_rules/sports-rules/029_age-of-empires.md) |
| SPORTS RULES | Apex Legends | `Apex Legends` | [030_apex-legends.md](../docs/04_sportsbook_rules/sports-rules/030_apex-legends.md) |
| SPORTS RULES | Arena of Valor | `arena-of-valor` | [031_arena-of-valor.md](../docs/04_sportsbook_rules/sports-rules/031_arena-of-valor.md) |
| SPORTS RULES | Artifact | `Artifact` | [032_artifact.md](../docs/04_sportsbook_rules/sports-rules/032_artifact.md) |
| SPORTS RULES | Brawl Stars | `brawl-stars` | [033_brawl-stars.md](../docs/04_sportsbook_rules/sports-rules/033_brawl-stars.md) |
| SPORTS RULES | PlayerUnknown's Battlegrounds (PUBG) | `pubg` | [034_pubg.md](../docs/04_sportsbook_rules/sports-rules/034_pubg.md) |
| SPORTS RULES | PlayerUnknown's Battlegrounds (PUBG ) Mobile | `pubg-mobile` | [035_pubg-mobile.md](../docs/04_sportsbook_rules/sports-rules/035_pubg-mobile.md) |
| SPORTS RULES | Deadlock | `deadlock` | [036_deadlock.md](../docs/04_sportsbook_rules/sports-rules/036_deadlock.md) |
| SPORTS RULES | Clash of Clans | `clash-of-clans` | [037_clash-of-clans.md](../docs/04_sportsbook_rules/sports-rules/037_clash-of-clans.md) |
| SPORTS RULES | Clash Royale | `clash-royale` | [038_clash-royale.md](../docs/04_sportsbook_rules/sports-rules/038_clash-royale.md) |
| SPORTS RULES | Cross Fire | `cross-fire` | [039_cross-fire.md](../docs/04_sportsbook_rules/sports-rules/039_cross-fire.md) |
| SPORTS RULES | Cross Fire HD league | `cross-fire-hd-league` | [040_cross-fire-hd-league.md](../docs/04_sportsbook_rules/sports-rules/040_cross-fire-hd-league.md) |
| SPORTS RULES | CrossFire Mobile | `crossFire-mobile` | [041_crossfire-mobile.md](../docs/04_sportsbook_rules/sports-rules/041_crossfire-mobile.md) |
| SPORTS RULES | Fortnite | `fortnite` | [042_fortnite.md](../docs/04_sportsbook_rules/sports-rules/042_fortnite.md) |
| SPORTS RULES | Free Fire | `free-fire` | [043_free-fire.md](../docs/04_sportsbook_rules/sports-rules/043_free-fire.md) |
| SPORTS RULES | Gwent | `gwent` | [044_gwent.md](../docs/04_sportsbook_rules/sports-rules/044_gwent.md) |
| SPORTS RULES | Gears of War | `gears-of-war` | [045_gears-of-war.md](../docs/04_sportsbook_rules/sports-rules/045_gears-of-war.md) |
| SPORTS RULES | Halo | `halo` | [046_halo.md](../docs/04_sportsbook_rules/sports-rules/046_halo.md) |
| SPORTS RULES | Heroes of Newerth | `Heroes-of-newerth` | [047_heroes-of-newerth.md](../docs/04_sportsbook_rules/sports-rules/047_heroes-of-newerth.md) |
| SPORTS RULES | League of Legends: Wild Rift | `league-of-legends-wild-rift` | [048_league-of-legends-wild-rift.md](../docs/04_sportsbook_rules/sports-rules/048_league-of-legends-wild-rift.md) |
| SPORTS RULES | World of Tanks | `world-of-tanks` | [049_world-of-tanks.md](../docs/04_sportsbook_rules/sports-rules/049_world-of-tanks.md) |
| SPORTS RULES | Mobile Legends | `mobile-legends` | [050_mobile-legends.md](../docs/04_sportsbook_rules/sports-rules/050_mobile-legends.md) |
| SPORTS RULES | Floorball | `floorball` | [051_floorball.md](../docs/04_sportsbook_rules/sports-rules/051_floorball.md) |
| SPORTS RULES | Football/Soccer | `football-soccer` | [052_football-soccer.md](../docs/04_sportsbook_rules/sports-rules/052_football-soccer.md) |
| SPORTS RULES | Mixed/Mythical Football | `mixed-mythical-football` | [053_mixed-mythical-football.md](../docs/04_sportsbook_rules/sports-rules/053_mixed-mythical-football.md) |
| SPORTS RULES | Futsal | `futsal` | [054_futsal.md](../docs/04_sportsbook_rules/sports-rules/054_futsal.md) |
| SPORTS RULES | Irish/GAA Sports (Gaelic Football/Hurling) | `irish-gga-sports` | [055_irish-gga-sports.md](../docs/04_sportsbook_rules/sports-rules/055_irish-gga-sports.md) |
| SPORTS RULES | Golf | `golf` | [056_golf.md](../docs/04_sportsbook_rules/sports-rules/056_golf.md) |
| SPORTS RULES | Greyhound Racing | `greyhound-racing ` | [057_greyhound-racing.md](../docs/04_sportsbook_rules/sports-rules/057_greyhound-racing.md) |
| SPORTS RULES | Handball | `handball` | [058_handball.md](../docs/04_sportsbook_rules/sports-rules/058_handball.md) |
| SPORTS RULES | Hockey (Non-Ice, including ’Field’, ‘Rink’ or ‘Inline’ Hockey). | `hockey-non-ice` | [059_hockey-non-ice.md](../docs/04_sportsbook_rules/sports-rules/059_hockey-non-ice.md) |
| SPORTS RULES | Horse Racing | `horse-racing` | [060_horse-racing.md](../docs/04_sportsbook_rules/sports-rules/060_horse-racing.md) |
| SPORTS RULES | Ice Hockey | `ice-hockey` | [061_ice-hockey.md](../docs/04_sportsbook_rules/sports-rules/061_ice-hockey.md) |
| SPORTS RULES | Motor Racing (Cars) | `motor-racing-cars` | [062_motor-racing-cars.md](../docs/04_sportsbook_rules/sports-rules/062_motor-racing-cars.md) |
| SPORTS RULES | Nascar/Busch Racing | `nascar-busch-racing` | [063_nascar-busch-racing.md](../docs/04_sportsbook_rules/sports-rules/063_nascar-busch-racing.md) |
| SPORTS RULES | Rally | `rally` | [064_rally.md](../docs/04_sportsbook_rules/sports-rules/064_rally.md) |
| SPORTS RULES | Motorbikes | `motorbikes` | [065_motorbikes.md](../docs/04_sportsbook_rules/sports-rules/065_motorbikes.md) |
| SPORTS RULES | Netball | `netball` | [066_netball.md](../docs/04_sportsbook_rules/sports-rules/066_netball.md) |
| SPORTS RULES | Olympics | `olympics` | [067_olympics.md](../docs/04_sportsbook_rules/sports-rules/067_olympics.md) |
| SPORTS RULES | Padel | `padel` | [068_padel.md](../docs/04_sportsbook_rules/sports-rules/068_padel.md) |
| SPORTS RULES | Poker | `poker` | [069_poker.md](../docs/04_sportsbook_rules/sports-rules/069_poker.md) |
| SPORTS RULES | Pool | `pool` | [070_pool.md](../docs/04_sportsbook_rules/sports-rules/070_pool.md) |
| SPORTS RULES | Rugby League | `rugby_league` | [071_rugby-league.md](../docs/04_sportsbook_rules/sports-rules/071_rugby-league.md) |
| SPORTS RULES | Rugby Union | `rugby_union` | [072_rugby-union.md](../docs/04_sportsbook_rules/sports-rules/072_rugby-union.md) |
| SPORTS RULES | Snooker | `snooker` | [073_snooker.md](../docs/04_sportsbook_rules/sports-rules/073_snooker.md) |
| SPORTS RULES | Speedway | `speedway` | [074_speedway.md](../docs/04_sportsbook_rules/sports-rules/074_speedway.md) |
| SPORTS RULES | Squash | `squash` | [075_squash.md](../docs/04_sportsbook_rules/sports-rules/075_squash.md) |
| SPORTS RULES | Table Tennis | `table-tennis` | [076_table-tennis.md](../docs/04_sportsbook_rules/sports-rules/076_table-tennis.md) |
| SPORTS RULES | Tennis | `tennis` | [077_tennis.md](../docs/04_sportsbook_rules/sports-rules/077_tennis.md) |
| SPORTS RULES | Volleyball | `volleyball` | [078_volleyball.md](../docs/04_sportsbook_rules/sports-rules/078_volleyball.md) |
| SPORTS RULES | Water Polo | `water-polo` | [079_water-polo.md](../docs/04_sportsbook_rules/sports-rules/079_water-polo.md) |
| SPORTS RULES | Winter Sports | `winter-sports` | [080_winter-sports.md](../docs/04_sportsbook_rules/sports-rules/080_winter-sports.md) |
| SPORTS RULES | Other Sports | `other-sports` | [081_other-sports.md](../docs/04_sportsbook_rules/sports-rules/081_other-sports.md) |

## 赔率与计算

| 模块 | 标题 | location | Markdown |
|---|---|---|---|
| ODDS CONVERSION | Odds Conversion | `root` | [001_odds-conversion.md](../docs/05_odds_math/odds-conversion/001_odds-conversion.md) |

## 投注风控服务 / BetGuard

| 模块 | 标题 | location | Markdown |
|---|---|---|---|
| BETGUARD | Introduction | `introduction` | [001_introduction.md](../docs/06_betguard_risk/betguard/001_introduction.md) |
| BETGUARD | Bet Placement process | `bet_placement_process` | [002_bet-placement-process.md](../docs/06_betguard_risk/betguard/002_bet-placement-process.md) |
| BETGUARD | Bet Placement Flowchart | `bet_placement_flowchart` | [003_bet-placement-flowchart.md](../docs/06_betguard_risk/betguard/003_bet-placement-flowchart.md) |
| BETGUARD | Resulting Bets & Reporting Selection Outcomes | `resulting_bets_reporting_selection_outcomes` | [004_resulting-bets-reporting-selection-outcomes.md](../docs/06_betguard_risk/betguard/004_resulting-bets-reporting-selection-outcomes.md) |
| BETGUARD | BetGuard Notes | `betGuard_notes` | [005_betguard-notes.md](../docs/06_betguard_risk/betguard/005_betguard-notes.md) |
| BETGUARD | AuthToken | `authToken` | [006_authtoken.md](../docs/06_betguard_risk/betguard/006_authtoken.md) |
| BETGUARD | Hash Parameter | `hash_parameter` | [007_hash-parameter.md](../docs/06_betguard_risk/betguard/007_hash-parameter.md) |
| BETGUARD | Hash Calculation | `hash_calculation` | [008_hash-calculation.md](../docs/06_betguard_risk/betguard/008_hash-calculation.md) |
| BETGUARD | Example | `example` | [009_example.md](../docs/06_betguard_risk/betguard/009_example.md) |
| BETGUARD | Partner API Security Checks | `partner_api_security_checks` | [010_partner-api-security-checks.md](../docs/06_betguard_risk/betguard/010_partner-api-security-checks.md) |
| BETGUARD | TS Parameter (Timestamp) | `ts_parameter_timestamp` | [011_ts-parameter-timestamp.md](../docs/06_betguard_risk/betguard/011_ts-parameter-timestamp.md) |
| BETGUARD | Hash Parameter | `partner_api_hash_parameter` | [012_partner-api-hash-parameter.md](../docs/06_betguard_risk/betguard/012_partner-api-hash-parameter.md) |
| BETGUARD | Hash Calculation | `partner_api_hash_calculation` | [013_partner-api-hash-calculation.md](../docs/06_betguard_risk/betguard/013_partner-api-hash-calculation.md) |
| BETGUARD | Hash Parameters Lists per Method | `hash_parameters_lists_per_method` | [014_hash-parameters-lists-per-method.md](../docs/06_betguard_risk/betguard/014_hash-parameters-lists-per-method.md) |
| BETGUARD | Example | `partner_api_example` | [015_partner-api-example.md](../docs/06_betguard_risk/betguard/015_partner-api-example.md) |
| BETGUARD | AuthToken; Binding with Currency | `authToken_binding_with_currency` | [016_authtoken-binding-with-currency.md](../docs/06_betguard_risk/betguard/016_authtoken-binding-with-currency.md) |
| BETGUARD | BetGuard API Calls | `betGuard_api` | [017_betguard-api.md](../docs/06_betguard_risk/betguard/017_betguard-api.md) |
| BETGUARD | CreateBet | `createBet` | [018_createbet.md](../docs/06_betguard_risk/betguard/018_createbet.md) |
| BETGUARD | CreateBet Request Sample | `createBet_request_sample` | [019_createbet-request-sample.md](../docs/06_betguard_risk/betguard/019_createbet-request-sample.md) |
| BETGUARD | CreateBet Response Sample | `createBet_response_sample` | [020_createbet-response-sample.md](../docs/06_betguard_risk/betguard/020_createbet-response-sample.md) |
| BETGUARD | GetMaxBetAmount | `get_max_bet_amount` | [021_get-max-bet-amount.md](../docs/06_betguard_risk/betguard/021_get-max-bet-amount.md) |
| BETGUARD | GetMaxBetAmount Request Sample | `get_max_bet_amount_request_sample` | [022_get-max-bet-amount-request-sample.md](../docs/06_betguard_risk/betguard/022_get-max-bet-amount-request-sample.md) |
| BETGUARD | GetMaxBetAmount Response Sample | `get_max_bet_amount_response_sample` | [023_get-max-bet-amount-response-sample.md](../docs/06_betguard_risk/betguard/023_get-max-bet-amount-response-sample.md) |
| BETGUARD | ResendFailedTransfers | `resend_failed_transfers` | [024_resend-failed-transfers.md](../docs/06_betguard_risk/betguard/024_resend-failed-transfers.md) |
| BETGUARD | ResendFailedTransfers Request Sample | `resend_failed_transfers_request_sample` | [025_resend-failed-transfers-request-sample.md](../docs/06_betguard_risk/betguard/025_resend-failed-transfers-request-sample.md) |
| BETGUARD | ResendFailedTransfers Response Sample | `resend_failed_transfers_response_sample` | [026_resend-failed-transfers-response-sample.md](../docs/06_betguard_risk/betguard/026_resend-failed-transfers-response-sample.md) |
| BETGUARD | MarkBetAsCashout | `mark_bet_as_cashout` | [027_mark-bet-as-cashout.md](../docs/06_betguard_risk/betguard/027_mark-bet-as-cashout.md) |
| BETGUARD | MarkBetAsCashout Request Sample | `mark_bet_as_cashout_request_sample` | [028_mark-bet-as-cashout-request-sample.md](../docs/06_betguard_risk/betguard/028_mark-bet-as-cashout-request-sample.md) |
| BETGUARD | MarkBetAsCashout Response Sample | `mark_bet_as_cashout_response_sample` | [029_mark-bet-as-cashout-response-sample.md](../docs/06_betguard_risk/betguard/029_mark-bet-as-cashout-response-sample.md) |
| BETGUARD | CheckAndMarkBetAsCashout | `check_and_mark_bet_as_cashout` | [030_check-and-mark-bet-as-cashout.md](../docs/06_betguard_risk/betguard/030_check-and-mark-bet-as-cashout.md) |
| BETGUARD | CheckAndMarkBetAsCashout Request Sample | `check_and_mark_bet_as_cashout_request_sample` | [031_check-and-mark-bet-as-cashout-request-sample.md](../docs/06_betguard_risk/betguard/031_check-and-mark-bet-as-cashout-request-sample.md) |
| BETGUARD | CheckAndMarkBetAsCashout Response Sample | `check_and_mark_bet_as_cashout_response_sample` | [032_check-and-mark-bet-as-cashout-response-sample.md](../docs/06_betguard_risk/betguard/032_check-and-mark-bet-as-cashout-response-sample.md) |
| BETGUARD | ReturnBet | `return_bet` | [033_return-bet.md](../docs/06_betguard_risk/betguard/033_return-bet.md) |
| BETGUARD | ReturnBet Request Sample | `return_bet_request_sample` | [034_return-bet-request-sample.md](../docs/06_betguard_risk/betguard/034_return-bet-request-sample.md) |
| BETGUARD | ReturnBet Response Sample | `return_bet_response_sample` | [035_return-bet-response-sample.md](../docs/06_betguard_risk/betguard/035_return-bet-response-sample.md) |
| BETGUARD | UpdateClient | `update_client` | [036_update-client.md](../docs/06_betguard_risk/betguard/036_update-client.md) |
| BETGUARD | UpdateClient Request Sample | `update_client_request_sample` | [037_update-client-request-sample.md](../docs/06_betguard_risk/betguard/037_update-client-request-sample.md) |
| BETGUARD | UpdateClient Response Sample | `update_client_response_sample` | [038_update-client-response-sample.md](../docs/06_betguard_risk/betguard/038_update-client-response-sample.md) |
| BETGUARD | GetClientDetails | `partner_api_get_client_details` | [039_partner-api-get-client-details.md](../docs/06_betguard_risk/betguard/039_partner-api-get-client-details.md) |
| BETGUARD | GetClientDetails Request Sample | `partner_api_request_sample` | [040_partner-api-request-sample.md](../docs/06_betguard_risk/betguard/040_partner-api-request-sample.md) |
| BETGUARD | GetClientDetails Response Sample | `partner_api_response_sample` | [041_partner-api-response-sample.md](../docs/06_betguard_risk/betguard/041_partner-api-response-sample.md) |
| BETGUARD | BetPlaced | `partner_api_bet_placed` | [042_partner-api-bet-placed.md](../docs/06_betguard_risk/betguard/042_partner-api-bet-placed.md) |
| BETGUARD | BetPlaced Request Sample | `bet_placed_request_sample` | [043_bet-placed-request-sample.md](../docs/06_betguard_risk/betguard/043_bet-placed-request-sample.md) |
| BETGUARD | BetPlaced Response Sample | `bet_placed_response_sample` | [044_bet-placed-response-sample.md](../docs/06_betguard_risk/betguard/044_bet-placed-response-sample.md) |
| BETGUARD | BetResulted | `bet_resulted` | [045_bet-resulted.md](../docs/06_betguard_risk/betguard/045_bet-resulted.md) |
| BETGUARD | BetResulted Request Sample | `bet_resulted_request_sample` | [046_bet-resulted-request-sample.md](../docs/06_betguard_risk/betguard/046_bet-resulted-request-sample.md) |
| BETGUARD | BetResulted Response Sample | `bet_resulted_response_sample` | [047_bet-resulted-response-sample.md](../docs/06_betguard_risk/betguard/047_bet-resulted-response-sample.md) |
| BETGUARD | BetResulted Retry Logic | `bet_resulted_retry_logic` | [048_bet-resulted-retry-logic.md](../docs/06_betguard_risk/betguard/048_bet-resulted-retry-logic.md) |
| BETGUARD | Rollback | `partner_api_rollback` | [049_partner-api-rollback.md](../docs/06_betguard_risk/betguard/049_partner-api-rollback.md) |
| BETGUARD | Rollback Request Sample | `rollback_request_sample` | [050_rollback-request-sample.md](../docs/06_betguard_risk/betguard/050_rollback-request-sample.md) |
| BETGUARD | Rollback Response Sample | `rollback_response_sample` | [051_rollback-response-sample.md](../docs/06_betguard_risk/betguard/051_rollback-response-sample.md) |
| BETGUARD | Rollback Retry Logic | `rollback_retry_logic` | [052_rollback-retry-logic.md](../docs/06_betguard_risk/betguard/052_rollback-retry-logic.md) |
| BETGUARD | Client | `partner_api_client` | [053_partner-api-client.md](../docs/06_betguard_risk/betguard/053_partner-api-client.md) |
| BETGUARD | Bet Selection | `partner_api_bet_selection` | [054_partner-api-bet-selection.md](../docs/06_betguard_risk/betguard/054_partner-api-bet-selection.md) |
| BETGUARD | Errors returned by the FeedConstruct | `errors_returned_by_the_feed_construct` | [055_errors-returned-by-the-feed-construct.md](../docs/06_betguard_risk/betguard/055_errors-returned-by-the-feed-construct.md) |
| BETGUARD | Errors returned by the Partner | `errors_returned_by_the_partner` | [056_errors-returned-by-the-partner.md](../docs/06_betguard_risk/betguard/056_errors-returned-by-the-partner.md) |
| BETGUARD | Bet Limits - Global Limits | `bet_limits_global_limits` | [057_bet-limits-global-limits.md](../docs/06_betguard_risk/betguard/057_bet-limits-global-limits.md) |
| BETGUARD | Client Default | `client_default` | [058_client-default.md](../docs/06_betguard_risk/betguard/058_client-default.md) |
| BETGUARD | Multiple Bets | `multiple_bets` | [059_multiple-bets.md](../docs/06_betguard_risk/betguard/059_multiple-bets.md) |
| BETGUARD | Reset Time and Global Live Delay | `reset_time_and_global_live_delay` | [060_reset-time-and-global-live-delay.md](../docs/06_betguard_risk/betguard/060_reset-time-and-global-live-delay.md) |
| BETGUARD | Bet Limits - Sport Limits | `bet_limits_sport_limits` | [061_bet-limits-sport-limits.md](../docs/06_betguard_risk/betguard/061_bet-limits-sport-limits.md) |
| BETGUARD | How to apply limits? | `how_to_apply_limits` | [062_how-to-apply-limits.md](../docs/06_betguard_risk/betguard/062_how-to-apply-limits.md) |
| BETGUARD | Currency Codes | `currency_codes` | [063_currency-codes.md](../docs/06_betguard_risk/betguard/063_currency-codes.md) |
| BETGUARD | .Net/C# | `cSharp` | [064_csharp.md](../docs/06_betguard_risk/betguard/064_csharp.md) |
| BETGUARD | Change Log | `betGuard_change_log` | [065_betguard-change-log.md](../docs/06_betguard_risk/betguard/065_betguard-change-log.md) |
