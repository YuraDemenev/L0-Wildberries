package handler

import (
	"L0/pkg/models"
	"L0/pkg/mynats"
	"L0/pkg/streaming"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// POST
func (h *Handler) postCreateProduct(c *gin.Context) {
	jsonText := c.Request.PostFormValue("inputJson")
	var order models.Order

	//Get Json from text field
	err := json.Unmarshal([]byte(jsonText), &order)
	if err != nil {
		logrus.Infof("ErrorStatus : %d |  Error : %s",
			http.StatusUnsupportedMediaType, err.Error())
		return
	}

	//Chec is data correct
	err = order.CheckData()
	if err != nil {
		logrus.Infof("ErrorStatus : %d |  Error : %s",
			http.StatusUnsupportedMediaType, err.Error())
		return
	}

	//Publish it
	err = streaming.PublishNats("test", jsonText)
	if err != nil {
		logrus.Infof("ErrorStatus : %d |  Error : %s",
			http.StatusInternalServerError, err.Error())
		return
	}

	return
}

func (h *Handler) postOrderInfo(c *gin.Context) {
	//Get order id
	orderId := c.Request.Header.Get("id")
	orderIdInt, _ := strconv.Atoi(orderId)

	//get order from cache
	order := mynats.Cache[orderIdInt]

	script := `<script>
	var jsonStringGo={{.}};
	var jsonString = JSON.stringify(jsonStringGo, null, 2)
	document.getElementById('inputJson').value = jsonString;
	</script>`

	tmpl, _ := template.New("t").Parse(script)

	tmpl.Execute(c.Writer, order)
}

// GET
func (h *Handler) getCreateProduct(c *gin.Context) {
	c.HTML(http.StatusOK, "CreateProduct.html", gin.H{})
}

func (h *Handler) getOrderInfo(c *gin.Context) {
	myButtonsSlice := make([]template.HTML, len(mynats.Cache))
	i := 0

	for _, orders := range mynats.Cache {
		divString := fmt.Sprintf(`
		<div>
			<button type='submit' class='btn btn-primary' hx-post='/order/GetInfo'
			hx-target='#errorsAlert' hx-headers='{"id":"%d"}'>
			%s
			</button>
		</div>  
		`, orders.Id, orders.TrackNumber)
		myButtonsSlice[i] = template.HTML(divString)
		i++
	}

	c.HTML(http.StatusOK, "GetOrder.html", gin.H{
		"myButtons": myButtonsSlice,
	})

}
