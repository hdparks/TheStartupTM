package main

import (
	"fmt"
	"math"
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
    cashPerSecond int
    users int
    usersPerSecondFromFeatures float64
    usersPerSecondFromMarketers float64
    usersPerSecondFromBugs float64
    features int
    featuresPerSecond float64
    bugs int
    bugsPerSecondPerFeature float64
    bugsPerSecondPerDev float64
    devs int
    qa int
    marketers int
    
    progressTowardFeature float64 
    progressTowardBug float64 
    progressTowardBugFix float64
    progressTowardUser float64 
    progressTowardLostUser float64
    copyPasteModifier int

    height int
    width int

    windowHeight int
    windowWidth int

    helpWindow bool
    helpModel help.Model

    
    cashParticles [20]particle
    cashParticlesVisible int

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
var FEATURE_SECONDS_PER_USER = 30
var USERS_PER_SECOND_PER_FEATURE = 1./30.
var USERS_PER_SECOND_PER_MARKERTER = 1./30.
var USERS_PER_SECOND_PER_BUG= 1./180.
var CASH_PER_SECOND_PER_USER_PER_FEATURE = 1
var BUGS_PER_SECOND_PER_FEATURE = 1./200.
var BUGS_PER_SECOND_PER_DEV = 1./200.
var BUGS_PER_SECOND_PER_QA = 1./60.
var FEATURES_PER_SECOND_PER_DEV = 1./60.
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
        cashPerSecond: 0,
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
        cashParticlesVisible: 0,

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

    m.featuresPerSecond = float64(m.devs) * FEATURES_PER_SECOND_PER_DEV
    m.progressTowardFeature += m.featuresPerSecond
    newFeatures := math.Floor(m.progressTowardFeature)
    m.features += int(newFeatures)
    m.progressTowardFeature -= newFeatures

    bugsFixedPerSecond := float64(m.qa) * BUGS_PER_SECOND_PER_QA
    m.progressTowardBugFix += bugsFixedPerSecond
    bugFixes := math.Floor(m.progressTowardBugFix)
    m.bugs -= int(bugFixes)
    m.progressTowardBugFix -= bugFixes 
    

    m.bugsPerSecondPerDev = float64(m.devs) * BUGS_PER_SECOND_PER_DEV
    m.bugsPerSecondPerFeature = float64(m.features) * BUGS_PER_SECOND_PER_FEATURE
    bugsPerSecond := m.bugsPerSecondPerDev + m.bugsPerSecondPerFeature
    m.progressTowardBug += bugsPerSecond
    newBugs := math.Floor(m.progressTowardBug)
    m.bugs += int(newBugs)
    m.progressTowardBug -= newBugs


    m.usersPerSecondFromFeatures = float64(m.features)* USERS_PER_SECOND_PER_FEATURE
    m.usersPerSecondFromMarketers = float64(m.marketers) * USERS_PER_SECOND_PER_MARKERTER
    usersAddedPerSecond :=  m.usersPerSecondFromFeatures + m.usersPerSecondFromMarketers
    m.progressTowardUser += usersAddedPerSecond
    newUsers := math.Floor(m.progressTowardUser)
    m.users += int(newUsers)
    m.progressTowardUser -= newUsers

    m.usersPerSecondFromBugs = float64(m.bugs) * USERS_PER_SECOND_PER_BUG
    usersLostPerSecond := m.usersPerSecondFromBugs
    m.progressTowardLostUser += usersLostPerSecond
    lostUsers := math.Floor(m.progressTowardLostUser)
    m.users -= int(lostUsers)
    m.progressTowardUser -= lostUsers

    
    m.cashPerSecond = CASH_PER_SECOND_PER_USER_PER_FEATURE * m.users * m.features
    m.cash += m.cashPerSecond
    m.cashParticlesVisible = min(int(math.Log2(float64(m.cashPerSecond))), len(m.cashParticles))

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
    HireDev key.Binding
    FireDev key.Binding
    HireQA key.Binding
    FireQA key.Binding
    HireMarketing key.Binding
    FireMarketing key.Binding
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
        {k.HireDev, k.FireDev, k.FocusBugs, k.FocusNewFeatures},
        {k.Help},
    }
}

var devKeys = devKeyMap{
    HireDev: key.NewBinding(
        key.WithKeys("h"),
        key.WithHelp("h","hire dev"),
    ),
    HireQA: key.NewBinding(
        key.WithKeys("y"),
        key.WithHelp("y","hire qa"),
    ),
    HireMarketing: key.NewBinding(
        key.WithKeys("t"),
        key.WithHelp("t","hire marketing"),
    ),
    FireDev: key.NewBinding(
        key.WithKeys("f"),
        key.WithHelp("f","fire dev"),
    ),
    FireQA: key.NewBinding(
        key.WithKeys("r"),
        key.WithHelp("r","fire qa"),
    ),
    FireMarketing: key.NewBinding(
        key.WithKeys("e"),
        key.WithHelp("e","fire marketing"),
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
        m.windowWidth = msg.Width
        m.windowHeight = msg.Height
        m.helpModel.Width = msg.Width

    case tea.KeyMsg: 

        if msg.String() == tea.KeyEsc.String(){
            return m, tea.Quit;
        }

        if (m.scene == Start){
            m.scene = Game
        }

        switch {

        case key.Matches(msg, devKeys.HireDev):
            m.devs += 1

        case key.Matches(msg, devKeys.HireQA):
            m.qa += 1

        case key.Matches(msg, devKeys.HireMarketing):
            m.marketers += 1

        case key.Matches(msg, devKeys.FireDev):
            m.devs -= 1

        case key.Matches(msg, devKeys.FireQA):
            m.qa -= 1

        case key.Matches(msg, devKeys.FireMarketing):
            m.marketers -= 1

        case key.Matches(msg, devKeys.FocusBugs):
            m.devFocus = max(0, m.devFocus - 1)

        case key.Matches(msg, devKeys.FocusNewFeatures):
            m.devFocus = min(10, m.devFocus + 1)
      
        case key.Matches(msg, devKeys.Help):
            m.helpWindow = !m.helpWindow

        case key.Matches(msg, devKeys.Features):
            m.progressTowardFeature = 1.

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
    
    style := startupStyle
    sec := time.Now().Unix()
    if (m.cash > 900000 && sec % 2 == 0){
        style = style.Foreground(lipgloss.Color("1"))
    } else {
        style = style.Foreground(lipgloss.Color("7"))
    }
    startupBuilding = style.Render(startupBuilding)

    s := cashTube + "\n" + startupBuilding

    l := float64(len(CashLevels))
    g := float64(CASH_CAP)
    y := math.Pow(g, 1/l)
    cashLog := math.Max(1.0,math.Log(float64(m.cash)))
    yLog := math.Log(y)
    cashSize := int(cashLog / yLog / 2)
    cashPile := CashLevels[cashSize]
    s = PlaceOverlay(0+cashPile.x, 5+cashPile.y, cashStyle.Render(cashPile.view), s, false)

    for i:=0;i<m.cashParticlesVisible;i++ {
        s = PlaceOverlay(m.cashParticles[i].x, 3 + m.cashParticles[i].y, cashStyle.Render(string(randomRune())), s, false)
    }

    return s 
}

func (m model) TableView() string {
   
    cols := []table.Column{
        {Title: "", Width: 12},
        {Title: "", Width: 8},
        {Title: "", Width: 16},
        {Title: "", Width: 16},
        {Title: "", Width: 16},
    }

    rows := []table.Row{
        {"Company Value", fmt.Sprintf("%v", m.pricePerShare)},
        {"Cash", fmt.Sprintf("%v", m.cash), fmt.Sprintf("$%d/sec",m.cashPerSecond)},
        {},
        {"Users", fmt.Sprintf("%v", m.users), fmt.Sprintf("%.2f/sec", m.usersPerSecondFromFeatures + m.usersPerSecondFromMarketers - m.usersPerSecondFromBugs)},
        {},
        {"Features", fmt.Sprintf("%v", m.features), fmt.Sprintf("%.2f Users/sec",m.usersPerSecondFromFeatures), fmt.Sprintf("%.2f Bugs/sec",m.bugsPerSecondPerFeature)},
        {"Bugs", fmt.Sprintf("%v", m.bugs), fmt.Sprintf("%.2f Users/sec", -m.usersPerSecondFromBugs)},
        {},
        {"Devs", fmt.Sprintf("%v",m.devs),fmt.Sprintf("%.2f Features/sec", m.featuresPerSecond), fmt.Sprintf("%.2f Bugs/sec",m.bugsPerSecondPerDev),fmt.Sprintf("%d $/sec", m.devs * DEV_SALARY_PER_SECOND)},
        {"QA", fmt.Sprintf("%v", m.qa)},
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

var centerStyle = lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center)

func (m model) View() string {
    viewStyle := centerStyle.Width(m.windowWidth).Height(m.windowHeight)
    switch(m.scene) {
    case Start:
        return viewStyle.Render(m.StartView())
    case Game:
        return viewStyle.Render(m.GameView())
    case End:
        return viewStyle.Render(m.EndView())
    }
    return "State not found"
}

func baseScreenStyle(m model) lipgloss.Style {
    style := baseStyle.Width(min(m.width,200)).Height(min(m.height,20)).UnsetAlign()
    return style
}


func (m model) GameView() string {
    style := baseScreenStyle(m)
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
    base += "\n--" + m.debug + "--\n"
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

