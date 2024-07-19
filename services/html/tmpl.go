package html

var tmpl = `
<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <link href="https://fonts.googleapis.com/css?family=Roboto:300,400&display=swap" rel="stylesheet">

    <link rel="stylesheet" href="fonts/icomoon/style.css">

    <link rel="stylesheet" href="css/owl.carousel.min.css">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="css/bootstrap.min.css">
    
    <!-- Style -->
    <link rel="stylesheet" href="css/style.css">

    <title>{{.Title}}</title>
  </head>
  <body>
  

  <div class="content">
    
    <div class="container">
      <h2 class="mb-5">{{.Title}}</h2>
      

      <div class="table-responsive custom-table-responsive">

        <table class="table custom-table">
          <thead>
            <tr>
				{{range .Columns}}
					{{if eq .IsCheckbox false}}
						<th scope="col">{{ .Key }}</th>
					{{else}}
						<th scope="col"></th>
					{{end}}
				{{else}}
					<div><strong>no rows</strong></div>
				{{end}}
            </tr>
          </thead>
          <tbody>
			{{- range .Rows}}
            <tr scope="row">
              {{range .}}<td>
				{{if eq .IsCheckbox false}}
                	{{if ne .Link ""}}<a href="{{.Link}}">{{end}}{{.Value}}
				{{else}}
					  <input type="checkbox" {{if eq .Value true}}checked=""{{end}} disabled="true"/>
				{{end}}
              {{end}}</td>
            </tr>
            <tr class="spacer"><td colspan="100"></td></tr>
			{{- end}}
          </tbody>
        </table>
      </div>


    </div>

  </div>
    
    

    <script src="js/jquery-3.3.1.min.js"></script>
    <script src="js/popper.min.js"></script>
    <script src="js/bootstrap.min.js"></script>
    <script src="js/main.js"></script>
  </body>
</html>`
