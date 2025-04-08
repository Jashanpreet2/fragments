package fragment

import (
	"encoding/json"
	"errors"
	"mime"
	"slices"
	"strings"
	"time"

	"github.com/Jashanpreet2/fragments/internal/logger"
	"github.com/Jashanpreet2/fragments/internal/utils"
)

type Fragment struct {
	Id           string    `json:"id" dynamodbav:"id"`
	OwnerId      string    `json:"ownerId" dynamodbav:"ownerId"`
	Created      time.Time `json:"created" dynamodbav:"created"`
	Updated      time.Time `json:"updated" dynamodbav:"updated"`
	FragmentType string    `json:"fragmentType" dynamodbav:"fragmentType"`
	Size         int       `json:"size" dynamodbav:"size"`
}

func (frag *Fragment) GetJson() (string, bool) {
	jsonData, err := json.Marshal(frag)
	if err != nil {
		return "", false
	}
	return string(jsonData), true
}

func (frag *Fragment) GetData() ([]byte, error) {
	file, err := ReadFragmentData(frag.OwnerId, frag.Id)
	if err != nil {
		logger.Sugar.Errorf("Failed to find data for the current fragment at userid: %s and fragment_id: %s", frag.OwnerId, frag.Id)
		return nil, err
	}
	return file, nil
}

// TODO: FIX THIS
func (frag *Fragment) SetData(data []byte) error {
	frag.Updated = time.Now()
	if err := WriteFragment(frag); err != nil {
		return err
	}
	if err := WriteFragmentData(frag.OwnerId, frag.Id, data); err != nil {
		return err
	}
	return nil
}

func (frag *Fragment) Save() error {
	return WriteFragment(frag)
}

func (frag *Fragment) MimeType() string {
	return frag.FragmentType
}

func (frag *Fragment) ConvertMimetype(ext string) ([]byte, string, error) {
	data, err := frag.GetData()
	mime.AddExtensionType(".md", "text/markdown")
	mime.AddExtensionType(".markdown", "text/markdown")
	mimeType := strings.Split(mime.TypeByExtension(ext), ";")[0]
	if mimeType == "" {
		return nil, "", errors.New("extension doesn't exist")
	}
	if err != nil {
		return nil, "", errors.New("unable to retrieve data")
	}
	if frag.MimeType() == "text/markdown" {
		logger.Sugar.Info(mimeType)
		if mimeType == "text/html" {
			return utils.ConvertMdToHtml(data), "text/markdown", nil
		}
	}
	return nil, "", errors.New("unsupported extension")
}

func (frag *Fragment) Formats() []string {
	mimeType := frag.MimeType()
	if mimeType == "text/md" || mimeType == "text/markdown" {
		return []string{"text/md", "text/markdown", "text/html"}
	}
	return []string{mimeType}
}

func GetUserFragmentIds(username string) ([]string, error) {
	return ListFragmentIDs(username)
}

func GetFragment(username string, fragment_id string) (*Fragment, error) {
	return ReadFragment(username, fragment_id)
}

func DeleteFragment(username string, fragment_id string) bool {
	return DeleteFragmentDB(username, fragment_id)
}

func IsSupportedType(typename string) bool {
	supportedTypes := []string{"text/1d-interleaved-parityfec", "text/cache-manifest", "text/calendar",
		"text/cql", "text/cql-expression", "text/cql-identifier", "text/css", "text/csv", "text/csv-schema",
		"text/directory", "text/dns", "text/ecmascript", "text/encaprtp", "text/enriched", "text/example",
		"text/fhirpath", "text/flexfec", "text/fwdred", "text/gff3", "text/grammar-ref-list", "text/hl7v2",
		"text/html", "text/javascript", "text/jcr-cnd", "text/markdown", "text/mizar", "text/n3", "text/parameters",
		"text/parityfec", "text/plain", "text/provenance-notation", "text/prs.fallenstein.rst", "text/prs.lines.tag",
		"text/prs.prop.logic", "text/prs.texi", "text/raptorfec", "text/RED", "text/rfc822-headers", "text/richtext",
		"text/rtf", "text/rtp-enc-aescm128", "text/rtploopback", "text/rtx", "text/SGML", "text/shaclc", "text/shex",
		"text/spdx", "text/strings", "text/t140", "text/tab-separated-values", "text/troff", "text/turtle", "text/ulpfec",
		"text/uri-list", "text/vcard", "text/vnd.a", "text/vnd.abc", "text/vnd.ascii-art", "text/vnd.curl", "text/vnd.debian.copyright",
		"text/vnd.DMClientScript", "text/vnd.dvb.subtitle", "text/vnd.esmertec.theme-descriptor", "text/vnd.exchangeable",
		"text/vnd.familysearch.gedcom", "text/vnd.ficlab.flt", "text/vnd.fly", "text/vnd.fmi.flexstor", "text/vnd.gml",
		"text/vnd.graphviz", "text/vnd.hans", "text/vnd.hgl", "text/vnd.in3d.3dml", "text/vnd.in3d.spot", "text/vnd.IPTC.NewsML",
		"text/vnd.IPTC.NITF", "text/vnd.latex-z", "text/vnd.motorola.reflex", "text/vnd.ms-mediapackage",
		"text/vnd.net2phone.commcenter.command", "text/vnd.radisys.msml-basic-layout", "text/vnd.senx.warpscript",
		"text/vnd.si.uricatalogue", "text/vnd.sun.j2me.app-descriptor", "text/vnd.sosi", "text/vnd.trolltech.linguist",
		"text/vnd.vcf", "text/vnd.wap.si", "text/vnd.wap.sl", "text/vnd.wap.wml", "text/vnd.wap.wmlscript", "text/vnd.zoo.kcl",
		"text/vtt", "text/wgsl", "text/xml", "text/xml-external-parsed-entity", "application/json"}

	return slices.Contains(supportedTypes, strings.Split(typename, ";")[0])
}
