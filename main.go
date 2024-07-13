package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/ansi"
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
    progressTowardBugFix int
    progressTowardUser int
    progressTowardLostUser int
    copyPasteModifier int

    height int
    width int

    helpWindow bool
    helpModel help.Model

    cashParticles [20]particle

    devFocus int
    devFocusProgress progress.Model
    
    gameTicking bool

    scene GameScene
    failureCause string

    // debug
    debug string
}

type GameScene int
const (
    Start GameScene = iota
    Game
    End
)

type particle struct {
    x int
    y int
}

var baseStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
var FEATURE_SECONDS_PER_USER = 60
var BUG_SECONDS_PER_LOST_USER = 180
var CASH_PER_SECOND_PER_USER_PER_FEATURE = 1
var DEV_FEATURE_SECONDS_PER_BUG = 200
var DEV_SECONDS_PER_FEATURE = 60
var DEV_SALARY_PER_SECOND = 1
var PRICE_PER_FEATURE = 100
var PRICE_PER_USER = 100
var PRICE_PER_BUG = -500
var PRICE_PER_DEV = 1000
var PRICE_PER_MARKETER = 1000
var CASH_CAP = 1000000


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
        progressTowardBugFix:0,
        progressTowardUser:0,
        progressTowardLostUser: 0,
        copyPasteModifier: 0,

        height: 0,
        width: 0,

        helpWindow: false,
        helpModel: help.New(),

        cashParticles: [20]particle{},

        devFocus: 10,
        devFocusProgress: progress.New(progress.WithSolidFill("4"), progress.WithWidth(10)),

        scene: Start,
        gameTicking: true,
    }
}


func (m model) Init() tea.Cmd {
    return tea.Batch(doGameTick(), doFrameTick(), textinput.Blink)
}

type GameTickMsg time.Time

func doGameTick() tea.Cmd {
    tick := tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return GameTickMsg(t)
    })
    return tick
}

func onGameTick(m model) model { 
    if (!m.gameTicking){
        return m
    }

    m.progressTowardFeature += (m.devs * m.devFocus) / 10
    newFeatures := m.progressTowardFeature / DEV_SECONDS_PER_FEATURE
    m.features += newFeatures
    m.progressTowardFeature -= newFeatures * DEV_SECONDS_PER_FEATURE 

    m.progressTowardBugFix += (m.devs * (10 - m.devFocus)) / 10
    bugFixes := m.progressTowardBugFix / DEV_SECONDS_PER_FEATURE
    m.bugs -= bugFixes
    m.progressTowardBugFix -= bugFixes * DEV_SECONDS_PER_FEATURE

    m.progressTowardBug += m.devs * m.features
    newBugs := m.progressTowardBug / DEV_FEATURE_SECONDS_PER_BUG
    m.bugs += newBugs
    m.progressTowardBug -= newBugs * DEV_FEATURE_SECONDS_PER_BUG

    m.progressTowardUser += m.features
    newUsers := m.progressTowardUser / FEATURE_SECONDS_PER_USER
    m.users += newUsers
    m.progressTowardUser -= newUsers * FEATURE_SECONDS_PER_USER

    m.progressTowardLostUser += m.bugs
    lostUsers := m.progressTowardLostUser / BUG_SECONDS_PER_LOST_USER 
    m.users -= lostUsers
    m.progressTowardUser -= lostUsers * BUG_SECONDS_PER_LOST_USER 

    
    m.cash += CASH_PER_SECOND_PER_USER_PER_FEATURE * m.users * m.features

    m.pricePerShare =   PRICE_PER_FEATURE * m.features +
                        PRICE_PER_DEV * m.devs +
                        PRICE_PER_BUG * m.bugs +
                        PRICE_PER_USER * m.users +
                        PRICE_PER_MARKETER * m.marketers
        
    if(m.cash > CASH_CAP){
        m.scene = End
        m.failureCause = "You've been crushed under the weight of your own success...\nA tragedy has befallen all mankind."
    }

    if(m.pricePerShare < 0) {
        m.scene = End
        m.failureCause = "Your enterprise has colapsed around you. A flash in the pan, nothing more."
    }

    return m
}




type FrameTickMsg time.Time
func doFrameTick() tea.Cmd {
    tick := tea.Tick(time.Second/12, func(t time.Time) tea.Msg {
        return FrameTickMsg(t)
    })
    return tick
}

func onFrameTick(m model) model {
    for i:=0;i<len(m.cashParticles);i++{
        d20 := rand.Intn(20)
        x := m.cashParticles[i].x
        if (d20 < x - 1){ m.cashParticles[i].x -= 1 }
        if (d20 > x + 1){ m.cashParticles[i].x += 1 }
        m.cashParticles[i].y = (m.cashParticles[i].y + rand.Intn(2)) % 20
    }
    return m
}


type devKeyMap struct {
    Hire key.Binding
    Fire key.Binding
    FocusBugs key.Binding
    FocusNewFeatures key.Binding
    Help key.Binding
    Features key.Binding
    Bugs key.Binding
}

func (k devKeyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Help}
}
func (k devKeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Hire, k.Fire, k.FocusBugs, k.FocusNewFeatures},
        {k.Help},
    }
}

var devKeys = devKeyMap{
    Hire: key.NewBinding(
        key.WithKeys("h"),
        key.WithHelp("h","hire dev"),
    ),
    Fire: key.NewBinding(
        key.WithKeys("f"),
        key.WithHelp("f","fire dev"),
    ),
    FocusBugs: key.NewBinding(
        key.WithKeys("b"),
        key.WithHelp("b","focus bugs"),
    ),
    FocusNewFeatures: key.NewBinding(
        key.WithKeys("n"),
        key.WithHelp("n","focus new features"),
    ),
    Features: key.NewBinding(
        key.WithKeys("{", "}","[","]","j","k"),
        key.WithHelp("jkl;","make features"),
    ),
    Bugs: key.NewBinding(
        key.WithKeys("1","2","3","4"),
        key.WithHelp("1234","fix bugs"),
    ),
    Help: key.NewBinding(
        key.WithKeys("?"),
        key.WithHelp("?", "help"),
    ),
}


func(m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = min(200,msg.Width)
        m.height = min(20,msg.Height)
        m.helpModel.Width = msg.Width

    case tea.KeyMsg: 
        m.debug = msg.String()

        if msg.String() == tea.KeyEsc.String(){
            return m, tea.Quit;
        }

        if (m.scene == Start){
            m.scene = Game
        }

        switch {

        case key.Matches(msg, devKeys.Hire):
            m.devs += 1

        case key.Matches(msg, devKeys.Fire):
            m.devs -= 1

        case key.Matches(msg, devKeys.FocusBugs):
            m.devFocus = max(0, m.devFocus - 1)

        case key.Matches(msg, devKeys.FocusNewFeatures):
            m.devFocus = min(10, m.devFocus + 1)
      
        case key.Matches(msg, devKeys.Help):
            m.helpWindow = !m.helpWindow

        case key.Matches(msg, devKeys.Features):
            m = m.ApplyProgressTowardFeature(1)

        case key.Matches(msg, devKeys.Bugs):
            m.bugs = max(0, m.bugs - 1)
        }


    case GameTickMsg:
        m = onGameTick(m)
        return m, doGameTick()

    case FrameTickMsg:
        m = onFrameTick(m)
        return m, doFrameTick()
    }

    return m, nil 
}

var cashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
var startupStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
var cashParticleRunes = []rune{'◜','\'',',','◝','◃','"','◟','◞'}
func randomRune() rune {
    return cashParticleRunes[rand.Intn(len(cashParticleRunes))]
}



func (m model) CashWindow() string {

    cashTube := "   |           |\n"
    cashTube += "  .─────────────.\n"
    cashTube += " /    Cash       \\\n"
    cashTube += "/.───────────────.\\\n"
    cashTube += "(                 )\n"
    cashTube += " `───────────────' \n"
    cashTube += strings.Repeat("\n", max(0,m.height - 14)) 


    cashTube = cashStyle.Render(cashTube)


    startupBuilding := "   ┌──────────┐   \n"
    startupBuilding += "┌─┬┴──────────┴┬─┐\n"
    startupBuilding += "│ │  STaRtupTM │ │\n"
    startupBuilding += "│ └────────────┘ │\n"
    startupBuilding += "│◫ ◫ ◫ ◫  ◫ ◫ ◫ ◫│\n"
    startupBuilding += "│◫ ◫ ◫ ◫  ◫ ◫ ◫ ◫│\n"
    startupBuilding += "└──────┮◚◚┭──────┘\n"
    startupBuilding = startupStyle.Render(startupBuilding)

    s := cashTube + "\n" + startupBuilding

    l := len(CashLevels)
    cashSize := min(m.cash / 1000, l-1)
    cashPile := CashLevels[cashSize]
    s = PlaceOverlay(0+cashPile.x, 5+cashPile.y, cashStyle.Render(cashPile.view), s, false)

    for i:=0;i<len(m.cashParticles);i++ {
        s = PlaceOverlay(m.cashParticles[i].x, 3 + m.cashParticles[i].y, cashStyle.Render(string(randomRune())), s, false)
    }

    return s 
}

func (m model) TableView() string {
   
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

    return t.View()
}


var devBorder = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63"));

func (m model) DevWindowView() string {
    dev := m.helpModel.FullHelpView(devKeys.FullHelp())
    return devBorder.Render(dev)
}

func maxWidth(s []string) int {
    maxW := 0
    for i:=0; i<len(s); i++ {
        maxW = max(ansi.PrintableRuneWidth(s[i]),maxW)
    }
    return maxW
}

func (m model) View() string {
    switch(m.scene) {
    case Start:
        return m.StartView()
    case Game:
        return m.GameView()
    case End:
        return m.EndView()
    }
    return "State not found"
}

func baseScreenStyle(m model) lipgloss.Style {
    style := baseStyle.Width(min(m.width,200)).Height(min(m.height,20)).UnsetAlign()
    return style
}


func (m model) GameView() string {
    style := baseScreenStyle(m)
    m.debug = fmt.Sprintf("%v",style.GetAlign())
    base := style.Render(m.TableView())

    cashView := m.CashWindow()
    
    w := maxWidth(strings.Split(cashView, "\n"))
    base = PlaceOverlay(m.width-w-1, 1, cashView, base, false)

    if (m.helpWindow) {
        devOverlay := m.DevWindowView()
        lines := strings.Split(devOverlay, "\n")
        width := maxWidth(lines)
        base = PlaceOverlay(m.width/2 - width/2, m.height/2 - len(lines)/2, devOverlay, base, true)
    }

    shortHelp := m.helpModel.ShortHelpView(devKeys.ShortHelp())
    base += "\n" + shortHelp + "\n"

    base += "\n" + m.debug + "\n"
    return base 
}

func (m model) StartView() string {
    style := baseScreenStyle(m)
    return style.Align(lipgloss.Center, lipgloss.Center).Render(`
████████╗██╗  ██╗███████╗    ███████╗████████╗ █████╗ ██████╗ ████████╗██╗   ██╗██████╗      ██╗████████╗███╗   ███╗██╗ 
╚══██╔══╝██║  ██║██╔════╝    ██╔════╝╚══██╔══╝██╔══██╗██╔══██╗╚══██╔══╝██║   ██║██╔══██╗    ██╔╝╚══██╔══╝████╗ ████║╚██╗
   ██║   ███████║█████╗      ███████╗   ██║   ███████║██████╔╝   ██║   ██║   ██║██████╔╝    ██║    ██║   ██╔████╔██║ ██║
   ██║   ██╔══██║██╔══╝      ╚════██║   ██║   ██╔══██║██╔══██╗   ██║   ██║   ██║██╔═══╝     ██║    ██║   ██║╚██╔╝██║ ██║
   ██║   ██║  ██║███████╗    ███████║   ██║   ██║  ██║██║  ██║   ██║   ╚██████╔╝██║         ╚██╗   ██║   ██║ ╚═╝ ██║██╔╝
   ╚═╝   ╚═╝  ╚═╝╚══════╝    ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝          ╚═╝   ╚═╝   ╚═╝     ╚═╝╚═╝ 
`)
}

func (m model) EndView() string {
    style := baseStyle
    base := style.Width(min(m.width,200)).Height(min(m.height,20)).Align(lipgloss.Center,lipgloss.Center).Render(m.failureCause)
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
