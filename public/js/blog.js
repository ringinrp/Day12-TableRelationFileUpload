let data = []

function addData(event) {

    event.preventDefault()

    let projectname = document.getElementById("project-name").value
    let description = document.getElementById("description").value
    let image = document.getElementById("image").files


    let gambar = URL.createObjectURL(image[0])
    console.log("gambar", image[0])
    console.log("gambar dengan path", gambar)

    let blog = {
        projectname,
        description,
        image,
        postAt: "21 November 2022",
        author: "ringin restu pati",
    }

    data.push(blog)
    console.log(data)

}

for (let index = 0; index < 3; index++) {
    document.getElementById("contents").innnerHTML = ``
}