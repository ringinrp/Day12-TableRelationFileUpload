function submitData() {
    // console.log("Berhasil !!")

    let name = document.getElementById("name").value
    let email = document.getElementById("email").value
    let phone = document.getElementById("phone").value
    let subject = document.getElementById("subject").value
    let message = document.getElementById("message").value
    console.log(name, email, phone, subject, message)

    // let emailReceiver = "ringinrp21@gmail.com"

    if (name === "") {
        return alert('Nama belum terisi')
    } else if (email === "") {
        return alert('Email belum terisi')
    } else if (phone === "") {
        return alert('Nomer telephone belum terisi')
    } else if (subject === "") {
        return alert('Subject belum terisi')
    } else if (message === "") {
        return alert('Message belum terisi')
    }

    let link = document.createElement(`a`)
    link.href = `mailto:${email}?subject=${subject}&body=Hallo nama saya ${name}, ${message}, silakan hubungi saya di ${phone}`

    link.click()

}