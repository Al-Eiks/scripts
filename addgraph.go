package main

import (
        "fmt"
        "os"
        "github.com/AlekSi/zabbix"
        "flag"
)

const (
        ZabbixServer string = "http://zabbix-infra.dedale.tf1.fr/zabbix"
)

//permet de récupérer l'ID d'un screen grâce à son nom
func GetScreenId(screenName string, api *zabbix.API)(screenId string) {
        res, _ := api.CheckScreen(screenName, zabbix.Params{"output": "extend"})
        return(res)
}

//permet de lister les éléments qui composent un screen
func GetScreenElem(screenName string, api *zabbix.API)(screenId []string) {
        res, _ := api.GetScreenElem(screenName, zabbix.Params{"output": "extend",
                                                              "selectScreenItems": "extend"})
        return(res)
}

//permet d'obtenir le nom d'un graphe d'après son ID
func GetGraphName(graphId string, api *zabbix.API)(graphName string) {
        res, _ := api.GetGraphName(zabbix.Params{"output": "extend",
                                                 "graphids": graphId})
        return(res)
}

//permet de lister les items supervisés sur un graphe
func GetGraphItems(graphId string, api *zabbix.API)(graphItems []string) {
        res, _ := api.GetGraphItems(graphId, zabbix.Params{"output": "extend",
                                                           "expandData": 1,
                                                           "graphids": graphId})
        return(res)
}

//permet d'obtenir la couleur d'un item supervisé sur un graphe précis
func GetGraphItemColor(graphId string, graphItemId string, api *zabbix.API)(graphItemColor string) {
        res, _ := api.GetGraphItemColor(graphItemId, zabbix.Params{"output": "extend",
                                                                   "expandData": 1,
                                                                   "graphids": graphId})
        return(res)
}

//permet de retourner l'ID d'un hôte selon son nom
func GetHostId(serverName string, api *zabbix.API)(hostID string) {
        res, _ := api.HostsGet(zabbix.Params{"output": "extend"})
        for _, server := range res {
                if server.Name == serverName {
                        hostID = server.HostId
                }
        }
        return
}

//permet de vérifier la présence d'un hôte d'après son ID sur un graphe
func CheckHostOnGraph(graphId string, hostId string, api *zabbix.API)(result bool) {
        res, _ := api.CheckHostPresence(hostId, zabbix.Params{"output": "extend",
                                                              "expandData": 1,
                                                              "graphids": graphId})
        return(res)
}

//permet de récupérer le nom des items supervisés sur un graphe
func GetItemKeyFromGraph(graphId string, api *zabbix.API)(itemKey string) {
        res, _ := api.GetItemKey(zabbix.Params{"output": "extend",
                                               "expandData": 1,
                                               "graphids": graphId})
        return(res)
}

//permet de récupéer sur un hôte l'ID d'un item d'après sa key
func GetItemId(key string, hostId string, api *zabbix.API)(itemId string) {
        res, _ := api.GetItemId(key, zabbix.Params{"output": "extend",
                                                   "hostids": hostId})
        return(res)
}

//Définition des flags qui serviront à l'exécution du script
var ptrUser = flag.String("user", "", "Zabbix username (account)")
var ptrPass = flag.String("password", "", "Zabbix password for user")
var ptrScreen = flag.String("screen", "", "Name of the screen concerned")
var ptrGraph = flag.String("graph", "", "Graph which will be modified")
var ptrGraphId = flag.String("graphid", "", "Graph which will be modified")
var ptrHost = flag.String("host", "", "Host to be added to the graph")

func main() {
        flag.Parse()
        if *ptrUser == "" || *ptrPass == "" || *ptrScreen == "" || ((*ptrGraph == "") && (*ptrGraphId == "")) || *ptrHost == ""{
                fmt.Println("Veuillez vérifier les paramètres \n-user & password \n-nom du screen \n-nom ou id du graph \n-nom de l'hôte")
                os.Exit(1)
        }
        //Connexion à l'API
        api := zabbix.NewAPI(ZabbixServer + "/api_jsonrpc.php")
        _, err := api.Login(*ptrUser, *ptrPass)
        if err != nil {
                fmt.Println("Erreur de login Zabbix", err.Error())
                os.Exit(1)
        }
        //Définition des couleurs utilisées par Zabbix
        colors := []string{"00C8C8", "C8C800", "C8C8C8", "009600", "960000", "000096", "960096", "009696", "969600",
                           "969696", "00FF00", "FF0000", "0000FF", "FF00FF", "00FFFF", "FFFF00", "FFFFFF", "00C800",
                           "C80000", "0000C8", "C800C8"}
        //Définition de la variable qui contiendra les items de graphe
        mGraphItems := []zabbix.GraphItem{}

        //Récupération de l'ID du screen défini en flag
        screenId := GetScreenId(*ptrScreen, api)
        if len(screenId) == 0  {
                //Si le screen n'existe pas, le programme s'arrête ici
                fmt.Println("Le screen indiqué n'existe pas")
                return
        } else {
                //Sinon on récupère la liste des ID des éléments qui composent le screen
                screenElements := GetScreenElem(*ptrScreen, api)
                //Pour tous les éléments du screen :
                for j, _ := range screenElements {
                        //On récupère leur nom d'après leur ID
                        graphNames := GetGraphName(screenElements[j], api)
                        //Si le nom du graphe défini en flag correspond à un graphe existant sur le screen  :
                        if graphNames == *ptrGraph || screenElements[j] == *ptrGraphId {
                                //On renomme la variable contenant l'ID du graphe qui sera modifié
                                graphId := screenElements[j]
                                //On récupère la liste des items qui composent le graphe en question
                                itemsId := GetGraphItems(graphId, api)
                                //Pour chaque item récupéré :
                                for i, _ := range itemsId {
                                        if itemsId[i] != "" {
                                                //On récupère sa couleur
                                                color := GetGraphItemColor(screenElements[j], itemsId[i], api)
                                                //On alimente l'array définie au début qui contiendra les items
                                                mGraphItems = append(mGraphItems, zabbix.GraphItem{color, itemsId[i]})
                                        }
                                }
                                //A ce niveau du script, ce dernier a seulement recrée une array contenant les items
                                //qui étaient déjà présents sur le graphe

                                //Maintenant on récupère l'ID de l'hôte pour lequel on veut rajouter l'item sur le graphe
                                hostId := GetHostId(*ptrHost, api)
                                //Si on récupère bien un ID et si l'hôte en question n'est pas déjà supervisé sur le graphe :
                                if hostId != "" && CheckHostOnGraph(graphId, hostId, api) != true {
                                        //Alors on récupère le nom de l'item concerné par le graphe
                                        itemKey := GetItemKeyFromGraph(graphId, api)
                                        //Puis on récupère l'ID correspondant à l'item en question pour l'hôte qu'on veut superviser
                                        itemId := GetItemId(itemKey, hostId, api)
                                        //On récupère la couleur du dernier item qui compose le graphe
                                        colorLastItem := GetGraphItemColor(screenElements[j], itemsId[len(itemsId)-1], api)
                                        //Pour la liste des couleurs
                                        for k, _ := range colors {
                                                //Tant que la dernière couleur ne correspond pas à une couleur de l'array,
                                                if colorLastItem != colors[k] {
                                                        //on continue de chercher
                                                        continue
                                                //Quand on a déterminé le "numéro" de la couleur
                                                } else {
                                                        var colorNewItem string
                                                        //Si celle-ci correspond à la dernière de la liste
                                                        if colorLastItem == colors[len(colors)-1] {
                                                                //Alors la couleur de notre prochain item correspondra à la couleur 1
                                                                colorNewItem = colors[0]
                                                        } else {
                                                                //Sinon sa couleur prend la prochaine valeur de la liste
                                                                colorNewItem = colors[k+1]
                                                        }
                                                        //On alimente l'array contenant les items du graphe avec notre nouvel item
                                                        mGraphItems = append(mGraphItems, zabbix.GraphItem{colorNewItem, itemId})
                                                }
                                        }
                                } else if CheckHostOnGraph(graphId, hostId, api) != false {
                                        fmt.Println("L'hôte est déjà supervisé sur le graphe")
                                        return
                                } else {
                                        fmt.Println("L'hôte indiqué n'existe pas dans la base de données")
                                        return
                                }
                                //On finit par envoyer la reqûete à l'API de Zabbix
                                api.Call("graph.update", zabbix.Params{"graphid": graphId,
                                                                       "gitems": mGraphItems})
                                fmt.Println("Graph updated")
                        } else  {
                                fmt.Println("Le graphe indiqué n'est pas présent sur le screen ou n'existe pas")
                                return
                        }
                }
        }
}
