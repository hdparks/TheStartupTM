package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
    pricePerShare int
    cash int
    users int
    features int
    bugs int
    devs int
    marketers int
    
    progressTowardFeature int
    progressTowardBug int
    progressTowardUser int
    progressTowardLostUser int
    copyPasteModifier int

    hireMode bool
    hireNum int

}

var baseStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
var FEATURE_SECONDS_PER_USER = 60
var CASH_PER_SECOND_PER_USER_PER_FEATURE = 1
var DEV_FEATURE_SECONDS_PER_BUG = 200
var DEV_SECONDS_PER_FEATURE = 60
var PRICE_PER_FEATURE = 100
var PRICE_PER_USER = 100
var PRICE_PER_BUG = -500
var PRICE_PER_DEV = 1000
var PRICE_PER_MARKETER = 1000

func initialModel() model {
    return model {
        pricePerShare: 0,
        cash: 0,
        users: 1,
        features: 0,
        bugs:0,
        devs:0,
        marketers:0,

        progressTowardFeature:0,
        progressTowardBug:0,
        progressTowardUser:0,
        progressTowardLostUser: 0,
        copyPasteModifier: 0,

        hireMode: false,
        hireNum: 1,
    }
}

type TickMsg time.Time

func (m model) Init() tea.Cmd {
    return tea.Batch(doTick(), textinput.Blink)
}

func doTick() tea.Cmd {
    tick := tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
    return tick
}

func onTick(m model) model { 

    m.progressTowardFeature += m.devs
    newFeatures := m.progressTowardFeature / DEV_SECONDS_PER_FEATURE
    m.features += newFeatures
    m.progressTowardFeature -= newFeatures * DEV_SECONDS_PER_FEATURE

    m.progressTowardBug += m.devs * m.features
    newBugs := m.progressTowardBug / DEV_FEATURE_SECONDS_PER_BUG
    m.bugs += newBugs
    m.progressTowardBug -= newBugs * DEV_FEATURE_SECONDS_PER_BUG

    m.progressTowardUser += m.features
    newUsers := m.progressTowardUser / FEATURE_SECONDS_PER_USER
    m.users += newUsers
    m.progressTowardUser -= newUsers * FEATURE_SECONDS_PER_USER
    
    m.cash += CASH_PER_SECOND_PER_USER_PER_FEATURE * m.users * m.features

    m.pricePerShare =   PRICE_PER_FEATURE * m.features +
                        PRICE_PER_DEV * m.devs +
                        PRICE_PER_BUG * m.bugs +
                        PRICE_PER_USER * m.users +
                        PRICE_PER_MARKETER * m.marketers
         
    return m
}

func(m model) UpdateMain(msg tea.KeyMsg) model {
    switch msg.String() {
    
    case "{", "}","[","]","j","k":
        m = m.ApplyProgressTowardFeature(1)

    case "ctrl+c", "ctrl-v":
        m.copyPasteModifier = min(m.copyPasteModifier+1,10)

    case "1","2":
        m.bugs = max(0, m.bugs - 1)

    case "h":
        m.hireMode = !m.hireMode
    }
    return m
}

func (m model) UpdateHire(msg tea.KeyMsg) model {
    switch msg.String() {

    case "h":
        m.hireMode = !m.hireMode

    case "+":
        m.hireNum += 1

    case "-":
        m.hireNum -= 1

    case "d":
        m.devs += m.hireNum

    }
    return m    
}

func(m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg: 
        if msg.String() == tea.KeyEsc.String() {
            fmt.Println("quitting")
            return m, tea.Quit;
        }

        if (m.hireMode) {
            m = m.UpdateMain(msg)
        } else {
            m = m.UpdateHire(msg)
        }

    case TickMsg:
        m = onTick(m)
        return m, doTick()

    }

    return m, nil 
}

func (m model) TableView() string {
    s := "The Startup:tm:\n\n"
   
    cols := []table.Column{
        {Title: "", Width: 16},
        {Title: "", Width: 16},
    }

    rows := []table.Row{
        {"Price per share", fmt.Sprintf("%v", m.pricePerShare)},
        {"Cash", fmt.Sprintf("%v", m.cash)},
        {"Users", fmt.Sprintf("%v", m.users)},
        {"Features", fmt.Sprintf("%v", m.features)},
        {"Bugs", fmt.Sprintf("%v", m.bugs)},
        {"Devs", fmt.Sprintf("%v", m.devs)},
        {"Marketers", fmt.Sprintf("%v", m.marketers)},
    }

    t := table.New(
        table.WithRows(rows),
        table.WithColumns(cols),
    )

    s += baseStyle.Render(t.View()) + "\n"

    return s
}

func (m model) HiringView() string {
    s := "The Startup:tm:\n\n"
    
    hiring := fmt.Sprintf("Hire %v devs\n",m.hireNum) 
    hiring += "\n"
    hiring += "(h to close)\n"

    return s + baseStyle.Render(hiring)
    
}

func (m model) View() string {
    
    base := m.TableView()

    if (m.hireMode) {
        hiringOverlay := m.HiringView()
        base = PlaceOverlay(10,5, hiringOverlay, base, true)
    }
    return base 
}

func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there has been an error: %v", err)
        os.Exit(1)
    }
}

func (m model) ApplyProgressTowardFeature(points int) model {
    m.progressTowardFeature += points
    for m.progressTowardFeature > m.features {
       m.progressTowardFeature -= m.features
       m.features += 1
    }

    return m
}
