package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func removeUnwantedChars(s string) string {
    return strings.Map(
        func(r rune) rune {
            //Barf..... This must be posible to do better, although i guess it's efficient
            if r == '0' || r == '1' || r == '2' || r == '3' || r == '4' || r == '5' || r == '6' || r == '7' || r == '8' || r == '9' || r == 'a' || r == 'b' || r == 'c' || r == 'd' || r == 'e' || r == 'f' {
                return r
            }
            return -1
        },
        s,
    )
}

func sanitizeMacAddressString(s string) string {
    s = strings.ToLower(s)
    s = removeUnwantedChars(s)
    return s
    
}

type MacBSTNode struct {
    left    *MacBSTNode
    right   *MacBSTNode
    OUI     string
    vendor  string

}

type MacBST struct {
    root *MacBSTNode
}

func checkerr(err any) {
    if err != nil {
        fmt.Errorf("Error %w", err)
    }
} 

func (t *MacBST) insert(oui string, vendor string) {
    if t.root == nil {
        t.root = &MacBSTNode{OUI: oui}
    } else {
        t.root.insert(oui, vendor)
    }

}

func (n *MacBSTNode) insert(oui string, vendor string) {
    if oui <= n.OUI {
        if n.left == nil {
            n.left = &MacBSTNode{OUI: oui, vendor: vendor}
        } else {
            n.left.insert(oui, vendor)
        }
    } else {
        if n.right == nil {
            n.right = &MacBSTNode{OUI: oui, vendor: vendor}
        } else {
            n.right.insert(oui, vendor)
        }
    }
}

func (t *MacBST) search(macaddress string) (string, string) {
    if t.root == nil {
        return "Error, root not set", "Error root not set"
    } else {
        return t.root.search(macaddress, "", "")
    }
}

func (n *MacBSTNode) search(macaddress string, candidateOUI string, candidateVendor string) (string, string) {
    //println("Searching in " + n.OUI + " " + n.vendor)
    if n.right == nil && n.left == nil {
        if strings.Contains(macaddress, n.OUI) {
            return n.OUI, n.vendor
        } else {
            return candidateOUI, candidateVendor
        }
        
    }
    if macaddress <= n.OUI {
        if n.left == nil {
            return n.OUI, n.vendor
        } else {
            return n.left.search(macaddress, candidateOUI, candidateVendor)
        }
    } else {
        if strings.Contains(macaddress, n.OUI) {
            candidateOUI = n.OUI
            candidateVendor = n.vendor
        }
        if n.right == nil {
            return n.OUI, n.vendor
        } else {
            return n.right.search(macaddress, candidateOUI, candidateVendor)
        }
    }
}

func loadOUIs(url string, bst *MacBST) {
    resp, err := http.Get(url)
    checkerr(err)
    bodybytes, err := io.ReadAll(resp.Body)
    checkerr(err)
    for _, line := range strings.Split(strings.TrimSuffix(string(bodybytes), "\n"), "\n") {
        ouisplit := strings.Split(line, ",")
        //println("OUI: " + sanitizeMacAddressString(ouisplit[1]) + " Vendor: " + ouisplit[2])
        bst.insert(sanitizeMacAddressString(ouisplit[1]), ouisplit[2])
    }

}

type MacResponse struct {
    Macaddress  string  `json:"mac"`
    OUI         string  `json:"OUI"`
    Vendor      string  `json:"Vendor"`
}
type APIError struct {
    Error       string  `json:"error"`
}

func MacLookupWrapper(bst *MacBST) gin.HandlerFunc {
    fn := func(c *gin.Context) {
        macstring := sanitizeMacAddressString(c.Query("mac"))
        oui, vendor := bst.search(macstring)
        //println("OUI: " + oui + " Vendor " + vendor + " Requested mac " + macstring)
        
        if strings.Contains(macstring, oui) && vendor != "" {
            c.JSON(http.StatusOK,MacResponse{Macaddress: sanitizeMacAddressString(macstring), OUI: oui, Vendor: vendor})
        } else {
            c.JSON(http.StatusNotFound, APIError{Error: "OUI not found, or not enough of mac address provided"})
        }
         
    
    }
    return gin.HandlerFunc(fn)
}


func main() {
    var OUIBST MacBST

    loadOUIs("http://standards-oui.ieee.org/oui/oui.csv", &OUIBST)
    loadOUIs("http://standards-oui.ieee.org/oui28/mam.csv", &OUIBST)
    loadOUIs("http://standards-oui.ieee.org/oui36/oui36.csv", &OUIBST)
    
    router := gin.Default()
    router.GET("/lookup", MacLookupWrapper(&OUIBST))

    router.Run("[::]:8080")
}