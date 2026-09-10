# Staff Operations Guide

This document is a guide for staff running the workshop.
On the day, facilitate the workshop in order while sharing [README.md](README.md) with participants. Have each team open [TEAM_GUIDE_en.md](TEAM_GUIDE_en.md) as well.
It contains the preparations, staff responsibilities on the day, and tutorial demonstration script that are not included in the participant guide.

To make this guide reusable at other conferences and meetups, the main text contains only venue-independent procedures.
Circumstances specific to each event are collected at the end in [Event-specific notes](#event-specific-notes). When an event is scheduled, read those notes together with the main text.

## Staffing

| Role | Guideline | Responsibilities |
|:---|:---|:---|
| Facilitator | 1 person | Share [README.md](README.md), follow the script, and keep time |
| Support staff | 1 per 2-3 teams | Guide team formation and circulate among groups during the exploration workshop |

The minimum staffing is one facilitator plus one support staff member, for a total of two people.

## Preparation (By the Day Before)

### Check the Venue and Equipment

- [ ] Confirm that participant Wi-Fi is available and can reach pkg.go.dev, GitHub, and Go Playground. At an unstable venue, prepare a backup connection such as tethering for at least the facilitator.
- [ ] Confirm the screen or projector for the facilitator. **Check what video input ports and adapters the venue provides.** Bring anything that is not available.
- [ ] Decide whether each team's PC will be connected to the projector during presentations. If teams will not switch connections, prepare a place to collect their work, such as a chat, and tell participants how to use it.
- [ ] Confirm the microphone type (handheld or lapel). If only a handheld microphone is available, decide whether screen operation and explanation should be split because the tutorial will require one hand to be occupied.
- [ ] Confirm the time staff can enter the venue and whether the organizer requires the tables to be rearranged. If so, work backward from the setup time to determine when staff should arrive.
- [ ] Confirm whether the layout must be restored afterward. If someone responsible for the next session will already be there, decide how to hand over the room.
- [ ] Confirm whether power is available for participants. If there is not enough for everyone, tell participants to charge their devices before arriving in the event announcement.
- [ ] Confirm how participants will reach the repository. If its URL will be registered in the event's materials field, **publish the repository before registering it**. If it is reflected before publication, the official site will lead participants to a 404 page.
- [ ] Confirm that participants have been told to bring a PC.
- [ ] Prepare a table layout with islands for approximately four people each.
- [ ] Confirm that Run and Share work in Go Playground and that teams can paste shared URLs into their Issue.

### Create Issues for Teams

This workshop is designed for **up to seven teams and 28 participants**.
Before the event, create one Issue for each team in this repository to use as its research log. If attendance is uncertain, create seven and close the unused ones on the day.

The Issue body is provided by the Issue template [Research log (workshop team)](../.github/ISSUE_TEMPLATE/research-log.md).
From the web interface, choose the template under "New issue" and replace the "Group X" in the title with the team's name.

To create seven Issues in one batch with the `gh` command, run:

```bash
for g in A B C D E F G; do
  gh issue create \
    --title "[Go探無比] Group ${g} research log" \
    --body "$(sed '1,/^---$/d' .github/ISSUE_TEMPLATE/research-log.md)"
done
```

Share the list of created Issue URLs with all staff before the event so it can be handed to participants verbally, on paper, or in chat once teams are assigned.

### Facilitator Rehearsal

- [ ] Read [README.md](README.md) aloud from beginning to end once.
- [ ] Practice the "Tutorial demonstration script" below by performing it. In particular, become comfortable searching the pkg.go.dev page with Ctrl+F, reaching the "Printing," "Other flags," and "Explicit argument indexes" sections, and running the code in Go Playground.
- [ ] Decide the browser zoom level for projection in advance. Judge it by whether the README table, scenario code blocks, and the pkg.go.dev in-page search dialog are readable from the back of the room. Record the chosen level in the [Event-specific notes](#event-specific-notes) so the next event can start there.

## Running the Workshop

### Tutorial Demonstration Script (00:05 - 00:25)

The facilitator demonstrates the following operations on screen.
Share the [tutorial README](00-tutorial/README_en.md) and run the live demonstration according to this script.
Have participants open the tutorial README and <https://pkg.go.dev/fmt> on their own devices and follow along (the tutorial section of the README already directs them there).

#### 1. Check the Scenario (2 minutes)

Open the [tutorial README](00-tutorial/README_en.md) and show the scenario code.

```go
fmt.Printf("%#[1]v %[1]T\n", value)
```

First run it in [Go Playground](https://go.dev/play/p/RTNSvn_p2Ai) and confirm the output together. Then press **Share** and demonstrate that the same code can be shared with the team through a URL.
Tell participants who have not finished setting up local Go that they can follow along in Playground.

> Message to convey: "Suppose you find this one line in a colleague's code: `%#[1]v`. What does it mean? Today, instead of starting with a web search, open the official documentation."

#### 2. Solve Question 1 in a Live Demonstration (8 minutes)

Open <https://pkg.go.dev/fmt> and perform the following steps.

1. The format-verb descriptions are in the Overview text. Use **Ctrl+F / Cmd+F** to search for `%v`, then find `%v`, `%#v`, and `%T` in the `General:` list under "The verbs:" in the "Printing" section.
2. Use Ctrl+F again to search for `Other flags`, and confirm that the effect of `#` differs by verb. Return to the `General:` list to establish the meaning of `%#v` itself.
3. Return to Go Playground and change `value` to a struct or map. Observe the difference between the `%v` and `%#v` output.

> Message to convey: "You do not need to read the documentation from top to bottom. Search for the symbol or keyword you need with Ctrl+F and read only the relevant section. Also, when you know the name of the function you want, press `f` to jump directly to its documentation."

#### 3. Solve Question 2 in a Live Demonstration (5 minutes)

Present the question, "Why is it printed twice even though there is only one argument?", and follow these steps.

1. In the Overview of <https://pkg.go.dev/fmt>, use Ctrl+F / Cmd+F to search for `[`. 
2. Read the "Explicit argument indexes" section and confirm that `[1]` specifies "use the first argument."

#### 4. Summarize the Research Pattern (3 minutes)

> Message to convey: "What we just did is the research pattern used throughout this workshop.
> (1) Open the official documentation -> (2) narrow it down to the relevant section with keywords -> (3) try it yourself and verify the behavior.
> The category READMEs collect reverse-search guidance for this pattern, such as [how to research 01-packages](01-packages/README.md). Return there first if you get stuck during the exploration workshop."

#### 5. Introduce the Scenario Structure (2 minutes)

Show that each scenario README already includes hints and answers in `<details>` (collapsible sections).
The answers include a **research route** explaining which primary sources to follow and in what order, so encourage participants to retrace the route themselves even after opening an answer.

### Guide Team Formation (00:25 - 00:30)

1. Following the facilitator's instructions, ask people to raise their hands for beginner, intermediate, or advanced, and count them.
2. Assign an area to each level and have participants move there.
3. Have participants form teams of approximately four in each area. Adjust groups to three to five if the numbers do not divide evenly.
4. Once teams are set, tell each team its name and the URL of its Issue.
5. **Explain the process verbally.** Assume participants will not read the documentation. It is enough to communicate these points:
   - Spend the first three minutes introducing yourselves.
   - One scenario per team; first decide on an entry point together.
   - Then split four people into two pairs and divide the questions.
   - Whoever finds something should comment on the Issue immediately.

### Circulate Among Groups During the Exploration Workshop (00:35 - 01:20)

Support staff circulate through their assigned areas and speak to teams from the following perspectives.

- **A team has stopped moving:** First bring them back to the entry points by asking, "Are there any sources in the scenario's investigation entry points that you have not opened yet?" If that does not help, guide them to the category README and its reverse-search procedure. The workshop's support policy is to show the research route, not give the answer directly.
- **A team is stuck on a scenario:** Remind them that they may open the `<details>` sections for hints and answers.
- **A research log has stopped:** Prompt them with, "What page are you looking at now? Comment that directly in the Issue."
- **One person in a pair has stopped:** Ask that person what they are looking at now. Assigning a different entry point often gets them moving again. If a team has not split into pairs, have them do so on the spot.
- **A team hesitates to open the answer:** If less than 20 minutes remain, tell them it is okay to open it. Add that the goal is to retrace the answer's research route.
- **A team has only a list of URLs:** Ask them to add what they were trying to verify, which search terms led them to the section, and what they learned there.
- **A topic is too broad and the team is losing focus:** Help them adjust the scope by asking, "If you had to narrow this to a question you can answer in the remaining time, what would it be?"

### Timekeeping

The facilitator is responsible for timekeeping. The following are likely pressure points and adjustment methods.

| Timing | Adjustment |
|:---|:---|
| The tutorial runs long | Skip the Question 2 slide and only introduce it by saying, "Search for `[` in the Overview to reach Explicit argument indexes." |
| Topic selection runs long | Direct participants to choose from the [scenario list](SCENARIOS_en.md) and stop selection at 00:35. |
| 20 minutes remain in the exploration workshop (01:00) | Address everyone and tell them it is okay to open the answers. |
| 10 minutes remain in the exploration workshop (01:10) | Address everyone and have them begin filling in the presentation summary. |
| Presentations run long | Reduce the planned number of presentations from three teams to two. Do not cut the closing. |

Do **not** visit every team for presentations; select approximately three. Choose from beginner, intermediate, and advanced, prioritizing interesting research discovered while circulating.
The time guideline is 3 teams x 2 minutes plus 4 minutes for closing, for a total of 10 minutes. The research logs of teams that do not present remain in the Issues, so tell participants that their work is still preserved.

Projecting <https://clock.shikakun.com/> helps participants see how much time remains.

### After the Closing

- [ ] Confirm that each team's Issue contains a presentation summary. If it does, anyone can review the results later.
- [ ] Confirm that the event label is applied to each Issue, such as `go-conference-2026`.
- [ ] Close unused backup Issues.

## Event-specific Notes

Add circumstances that depend on the venue or event rules here for each event.
Keeping them separate from the general procedures helps the next person distinguish what is venue-specific from what remains the same each time.

### Go Conference 2026 (2026-09-11 13:30-15:00 / 90-minute slot)

There are 22 participants (capacity: 28) and three staff members.
The following constraints are specific to this event and are based on the speaker manual distributed by the organizers. Treat them as overrides to the procedures in the main text.

#### Venue and Setup

- Staff can enter only from 20 minutes before the start. Speakers are to assemble 10 minutes before the start.
- Staff must rearrange the tables themselves. Assign setup responsibilities and arrival times in advance because islands must be created during the pre-event setup window.
- After the event, restore the designated standard layout, with separate specifications for rooms 2A and 2B. If the person responsible for the next session arrives first, the outgoing and incoming staff should work together to change the room to the next requested layout.

#### Equipment

- Projection uses the speaker's PC over HDMI. The venue does not provide adapters, so bring one if the PC does not have an HDMI port.
- There is no time for a pre-event rehearsal at the venue. Check the connection after entering, during the 20 minutes before the start.
- The internal rehearsal will be remote, so the actual projector and microphone cannot be tested. Decide the browser zoom level by projecting it at another venue on the previous day.
- Confirmed zoom level: (to be filled in after the previous day's check)
- Only a handheld microphone is available; there is no lapel microphone. Review the script with the assumption that one hand will be occupied while operating the screen during the tutorial.
- Participant and speaker Wi-Fi use separate networks. The organizers warn that the connection may be unstable, so a mobile Wi-Fi device has been arranged as a backup connection.
- There will not be enough outlets for all participants. The connpass event page will ask participants to charge their devices in advance, and power strips have been arranged for loan.

#### Schedule

- Venue staff will also give a signal before the end (the method will be announced on the day). Coordinate on the day so it does not duplicate the README's "10 minutes remaining" announcement.
- The workshop will not be streamed or archived.
- Official hashtag: `#gocon26`.
